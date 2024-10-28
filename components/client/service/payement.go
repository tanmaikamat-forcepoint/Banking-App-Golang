package service

import (
	"bankManagement/models/beneficiary"
	"bankManagement/models/client"
	"bankManagement/models/payments"
	"bankManagement/repository"
	"bankManagement/utils/log"
	"errors"

	"github.com/jinzhu/gorm"
)

type PaymentService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.WebLogger
}

func NewPaymentService(
	DB *gorm.DB,
	repository repository.Repository,
	log log.WebLogger,
) *PaymentService {
	return &PaymentService{
		DB,
		repository,
		log,
	}
}

func (srv *PaymentService) CreateBeneficiary(bfciary *beneficiary.Beneficiary) error {
	uow := repository.NewUnitOfWork(srv.DB)
	defer uow.RollBack()
	var tempBeneficiaryToCheck = beneficiary.Beneficiary{}
	srv.repository.GetFirstWhere(uow, &tempBeneficiaryToCheck, "client_id=?", bfciary.ClientID, "beneficiary_receiver_id=?", bfciary.BeneficiaryReceiverID)
	if tempBeneficiaryToCheck.ID != 0 {
		return errors.New("Beneficiary Already Exist")
	}
	tempSender := client.Client{}
	err := srv.repository.GetByID(uow, &tempSender, bfciary.ClientID)
	if err != nil {
		return err
	}
	tempReceiver := client.Client{}
	err = srv.repository.GetByID(uow, &tempReceiver, bfciary.BeneficiaryReceiverID)
	if err != nil {
		return err
	}
	err = srv.repository.Add(uow, bfciary)
	if err != nil {
		return err
	}
	uow.Commit()
	return nil

}

// func (srv *PaymentService) CreateBeneficiary1(beneficiary *beneficiary.Beneficiary) error {
// 	uow := repository.NewUnitOfWork(srv.DB)
// 	defer uow.RollBack()

// 	uow.Commit()
// 	return nil

// }

func (srv *PaymentService) GetAllBeneficiariesForClient(clientId uint, beneficiaries *[]beneficiary.Beneficiary) error {
	uow := repository.NewUnitOfWork(srv.DB)
	defer uow.RollBack()
	tempSender := client.Client{}
	err := srv.repository.GetByID(uow, &tempSender, clientId)
	if err != nil {
		return err
	}
	err = srv.repository.GetAll(uow, beneficiaries,
		srv.repository.Filter("client_id=?", clientId),
	)
	if err != nil {
		return err
	}

	uow.Commit()
	return nil

}

func (srv *PaymentService) DeleteBeneficiaryById(clientId uint, beneficiaryId uint) error {
	uow := repository.NewUnitOfWork(srv.DB)
	defer uow.RollBack()
	tempSender := client.Client{}
	err := srv.repository.GetByID(uow, &tempSender, clientId)
	if err != nil {
		return err
	}
	var tempBeneficiaryToCheck = beneficiary.Beneficiary{}
	err = srv.repository.GetFirstWhere(uow, &tempBeneficiaryToCheck, "client_id=?", clientId, "id=?", beneficiaryId)
	if tempBeneficiaryToCheck.ID == 0 || err != nil {
		srv.log.Error(err)
		return errors.New("Beneficiary Donot Exist")
	}

	err = srv.repository.DeleteById(uow, tempBeneficiaryToCheck, beneficiaryId)
	if err != nil {

		return err
	}

	uow.Commit()
	return nil

}

func (srv *PaymentService) CreatePaymentRequest(clientId uint, createdByUserId uint, paymentRequest *payments.PaymentRequestDTO) error {
	uow := repository.NewUnitOfWork(srv.DB)
	defer uow.RollBack()

	tempPaymentRequest := payments.PaymentRequest{
		Amount:         paymentRequest.Amount,
		SenderClientID: clientId,
	}
	tempSender := client.Client{}
	err := srv.repository.GetByID(uow, &tempSender, clientId)
	if err != nil {
		return err
	}

	if tempSender.Balance < paymentRequest.Amount {
		return errors.New("Insufficient Balance to Transact")
	}

	// tempBeneficiary := beneficiary.Beneficiary{}
	var tempBeneficiaryToCheck = beneficiary.Beneficiary{}
	err = srv.repository.GetByID(uow, &tempBeneficiaryToCheck, paymentRequest.BeneficiaryId)
	if tempBeneficiaryToCheck.ID == 0 || err != nil {
		srv.log.Error(err)
		return errors.New("Beneficiary Not found")
	}
	tempPaymentRequest.ReceiverClientID = tempBeneficiaryToCheck.BeneficiaryReceiverID
	tempPaymentRequest.AuthorizerBankId = tempSender.BankID
	tempPaymentRequest.CreatedByUserId = createdByUserId

	err = srv.repository.Add(uow, &tempPaymentRequest)

	if err != nil {
		return err
	}
	//Balance not deducted as payment is not accepted yet
	uow.Commit()
	return nil

}

func (srv *PaymentService) GetAllPaymentRequest(clientId uint, requests *[]payments.PaymentRequest) error {
	uow := repository.NewUnitOfWork(srv.DB)
	defer uow.RollBack()

	tempSender := client.Client{}
	err := srv.repository.GetByID(uow, &tempSender, clientId)
	if err != nil {
		return err
	}

	err = srv.repository.GetAll(uow, requests,
		srv.repository.Preload("ReceiverClient"),
		srv.repository.Filter("sender_client_id=?", clientId),
	)

	if err != nil {
		return err
	}
	//Balance not deducted as payment is not accepted yet
	uow.Commit()
	return nil

}

func (srv *PaymentService) GetAllPaymentsProcessed(clientId uint, requests *[]payments.Payment) error {
	uow := repository.NewUnitOfWork(srv.DB)
	defer uow.RollBack()

	tempSender := client.Client{}
	err := srv.repository.GetByID(uow, &tempSender, clientId)
	if err != nil {
		return err
	}

	err = srv.repository.GetAll(uow, requests,
		srv.repository.Preload("clients"),
		srv.repository.Filter("sender_client_id=?", clientId),
	)

	if err != nil {
		return err
	}
	//Balance not deducted as payment is not accepted yet
	uow.Commit()
	return nil

}
