package service

import (
	"bankManagement/models/client"
	"bankManagement/models/employee"
	"bankManagement/models/payments"
	"bankManagement/models/reports"
	"bankManagement/models/salaryDisbursement"
	"bankManagement/models/transaction"
	"bankManagement/repository"
	"bankManagement/utils/log"
	"errors"

	"github.com/jinzhu/gorm"
)

type ClientService struct {
	DB         *gorm.DB
	repository repository.Repository
	log        log.WebLogger
}

func NewClientService(
	DB *gorm.DB,
	repository repository.Repository,
	log log.WebLogger,
) *ClientService {
	return &ClientService{
		DB,
		repository,
		log,
	}
}

func (service *ClientService) CreateEmployee(emp *employee.Employee) error {
	uow := repository.NewUnitOfWork(service.DB)
	defer uow.RollBack()
	var tempClient = &client.Client{}
	err := service.repository.GetByID(uow, tempClient, emp.ClientID)
	if err != nil || tempClient.ID == 0 {
		return errors.New("Client Not found")
	}
	err = service.repository.Add(uow, emp)
	if err != nil {
		return err
	}
	uow.Commit()
	return nil

}

func (service *ClientService) Update(emp *employee.Employee) error {
	uow := repository.NewUnitOfWork(service.DB)
	defer uow.RollBack()
	var tempEmployee = &employee.Employee{}
	err := service.repository.GetByID(uow, tempEmployee, emp.ID)
	if err != nil || tempEmployee.ID == 0 {
		return errors.New("Employee Not found")
	}
	err = service.repository.Update(uow, emp)
	if err != nil {
		return err
	}
	uow.Commit()
	return nil

}

func (service *ClientService) GetAllEmployeesByClientId(emp *[]employee.Employee, clientId uint) error {
	uow := repository.NewUnitOfWork(service.DB)
	defer uow.RollBack()
	err := service.repository.GetAll(uow, emp,
		service.repository.Filter("client_id=?", clientId),
	)
	if err != nil {
		return err
	}

	uow.Commit()
	return nil

}

func (service *ClientService) DeleteEmployeeById(clientId uint) error {
	uow := repository.NewUnitOfWork(service.DB)
	defer uow.RollBack()
	tempEmployee := &employee.Employee{}
	err := service.repository.GetByID(uow, tempEmployee, clientId)

	if err != nil {
		return err
	}
	err = service.repository.DeleteById(uow, tempEmployee, clientId)

	if err != nil {
		return err
	}
	uow.Commit()
	return nil
}

func (service *ClientService) DisburseSalaryToOneEmployee(clientId uint, empId uint, approvedByUserId uint) error {
	uow := repository.NewUnitOfWork(service.DB)
	defer uow.RollBack()
	tempDisbursement := &salaryDisbursement.SalaryDisbursement{
		ClientID:        clientId,
		EmpID:           empId,
		CreatedByUserId: approvedByUserId,
	}
	tempClient := &client.Client{}
	err := service.repository.GetByID(uow, tempClient, clientId)
	if err != nil {
		return err
	}

	tempEmployeeDetails := &employee.Employee{}
	err = service.repository.GetByID(uow, tempEmployeeDetails, empId)
	if err != nil {
		return err
	}
	if tempClient.Balance < tempEmployeeDetails.SalaryAmount {
		return errors.New("Insufficient Balance")
	}
	tempClient.Balance = tempClient.Balance - tempEmployeeDetails.SalaryAmount
	service.repository.Update(uow, tempClient)

	tempDisbursement.SalaryAmount = tempEmployeeDetails.SalaryAmount
	//Can use api to disburse salary to account number. In that case status will be pending. and can be updated using webhooks
	tempTransaction := &transaction.Transaction{
		ClientID:          clientId,
		PaymentType:       "Transfer",
		TransactionType:   transaction.TransactionDebit,
		TransactionAmount: tempEmployeeDetails.SalaryAmount,
		TransactionStatus: transaction.TransactionStatusApproved,
	}
	err = service.repository.Add(uow, tempTransaction)

	if err != nil {
		return err
	}
	tempEmployeeDetails.TotalSalaryReceived += tempTransaction.TransactionAmount
	err = service.repository.Update(uow, tempEmployeeDetails)

	if err != nil {
		return err
	}
	tempDisbursement.TransactionID = tempTransaction.ID
	tempDisbursement.Status = salaryDisbursement.DisbursementStatusApproved

	err = service.repository.Add(uow, tempDisbursement)

	if err != nil {
		return err
	}
	uow.Commit()
	return nil
}

func (service *ClientService) DisburseSalaryAllEmployees(clientId uint, approvedByUserId uint) error {
	uow := repository.NewUnitOfWork(service.DB)
	defer uow.RollBack()
	var allEmployees []employee.Employee
	err := service.repository.GetAll(uow, &allEmployees,
		service.repository.Filter("client_id=?", clientId),
	)
	if err != nil {
		return err
	}

	tempClient := &client.Client{}
	err = service.repository.GetByID(uow, tempClient, clientId)
	if err != nil {
		return err
	}
	var amountToBeDisbursed float64 = 0
	for _, emp := range allEmployees {
		amountToBeDisbursed += emp.SalaryAmount
	}

	// tempEmployeeDetails := &employee.Employee{}
	// err = service.repository.GetByID(uow, tempEmployeeDetails, empId)
	// if err != nil {
	// 	return err
	// }
	if tempClient.Balance < amountToBeDisbursed {
		return errors.New("Insufficient Balance")
	}
	tempClient.Balance = tempClient.Balance - amountToBeDisbursed
	service.repository.Update(uow, tempClient)

	for _, emp := range allEmployees {
		tempDisbursement := &salaryDisbursement.SalaryDisbursement{
			ClientID:        clientId,
			EmpID:           emp.ID,
			CreatedByUserId: approvedByUserId,
		}
		tempDisbursement.SalaryAmount = emp.SalaryAmount
		// //Can use api to disburse salary to account number. In that case status will be pending. and can be updated using webhooks
		tempTransaction := &transaction.Transaction{
			ClientID:          clientId,
			PaymentType:       "Transfer",
			TransactionType:   transaction.TransactionDebit,
			TransactionAmount: emp.SalaryAmount,
			TransactionStatus: transaction.TransactionStatusApproved,
		}
		err = service.repository.Add(uow, tempTransaction)

		if err != nil {
			return err
		}
		emp.TotalSalaryReceived += tempTransaction.TransactionAmount
		err = service.repository.Update(uow, emp)

		if err != nil {
			return err
		}
		tempDisbursement.TransactionID = tempTransaction.ID
		tempDisbursement.Status = salaryDisbursement.DisbursementStatusApproved

		err = service.repository.Add(uow, tempDisbursement)

		if err != nil {
			return err
		}

	}

	uow.Commit()
	return nil
}

func (service *ClientService) GetSalaryReport(clientId uint, report *reports.SalaryReport) error {
	uow := repository.NewUnitOfWork(service.DB)
	defer uow.RollBack()
	var allEmployees []employee.Employee
	employeeCount := 0
	err := service.repository.GetAll(uow, &allEmployees,
		service.repository.Filter("client_id=?", clientId),
		service.repository.Count(100000, 0, &employeeCount),
	)
	if err != nil {
		return err
	}

	tempClient := &client.Client{}
	err = service.repository.GetByID(uow, tempClient, clientId)
	if err != nil {
		return err
	}
	var amountToBeDisbursed float64 = 0
	for _, emp := range allEmployees {
		amountToBeDisbursed += emp.SalaryAmount
	}

	SalaryReport := &reports.SalaryReport{
		ExpectedMonthlySalaryDisbursal: 0,
		TotalSalaryDisbursed:           0,
		AverageSalary:                  0,
		TotalEmployees:                 employeeCount,
		ClientID:                       clientId,
		ClientName:                     tempClient.ClientName,
	}
	// var disbursal struct {
	// 	sum float64
	// }
	// err = service.repository.Raw(uow, &disbursal, "Select SUM(salary_amount) as sum From employees where client_id=? ;", clientId)
	// if err != nil {
	// 	return err
	// }
	// SalaryReport.ExpectedMonthlySalaryDisbursal = disbursal.sum
	// SalaryReport.AverageSalary = SalaryReport.ExpectedMonthlySalaryDisbursal / float64(SalaryReport.TotalEmployees)
	var empDisbursement []reports.EmployeePaymentDTO
	for _, emp := range allEmployees {
		empDisbursement = append(empDisbursement, reports.EmployeePaymentDTO{
			EmpId:           emp.ID,
			SalaryDisbursed: emp.TotalSalaryReceived,
			MonthlySalary:   emp.SalaryAmount,
		})
		SalaryReport.TotalSalaryDisbursed += emp.TotalSalaryReceived
		SalaryReport.ExpectedMonthlySalaryDisbursal += emp.SalaryAmount
	}
	if len(allEmployees) > 0 {
		SalaryReport.AverageSalary = SalaryReport.ExpectedMonthlySalaryDisbursal / float64(SalaryReport.TotalEmployees)
	}
	SalaryReport.EmployeeDisbursementData = empDisbursement
	*report = *SalaryReport

	uow.Commit()
	return nil
}

func (service *ClientService) GetPaymentReport(clientId uint, report *reports.PaymentReport) error {
	uow := repository.NewUnitOfWork(service.DB)
	defer uow.RollBack()
	var allPaymentRequest []payments.PaymentRequest
	employeeCount := 0
	err := service.repository.GetAll(uow, &allPaymentRequest,
		service.repository.Filter("sender_client_id=? OR  receiver_client_id=?", clientId, clientId),
		service.repository.Count(100000, 0, &employeeCount),
	)
	if err != nil {
		return err
	}

	tempClient := &client.Client{}
	err = service.repository.GetByID(uow, tempClient, clientId)
	if err != nil {
		return err
	}

	PaymentReport := &reports.PaymentReport{
		AveragePaymentValue:       0,
		TotalPaymentsSent:         0,
		TotalPaymentsReceived:     0,
		TotalPaymentReceivedValue: 0,
		TotalPaymentSentValue:     0,
		TotalPaymentRequests:      0,
		ApprovedPaymentRequests:   0,
		RejectedPaymentRequests:   0,

		ClientID: clientId,
		// ClientName: tempClient.ClientName,
	}

	// var paymentCreated map[uint]reports.ClientPaymentDTO
	paymentCreated := make(map[uint]reports.ClientPaymentDTO)
	for _, req := range allPaymentRequest {
		if req.SenderClientID == clientId && req.Status == transaction.TransactionStatusApproved {
			//Sender is client . So money is received by other client
			receiver, ok := paymentCreated[req.ReceiverClientID]
			if !ok {
				paymentCreated[req.ReceiverClientID] = reports.ClientPaymentDTO{
					ClientID:                         req.ReceiverClientID,
					TotalPaymentSentByThisClient:     0,
					TotalPaymentReceivedByThisClient: 0,
				}
				receiver = paymentCreated[req.ReceiverClientID]
			}
			receiver.TotalPaymentReceivedByThisClient += req.PaymentAmount
			PaymentReport.TotalPaymentSentValue += req.PaymentAmount
			PaymentReport.TotalPaymentsSent += 1
			paymentCreated[req.ReceiverClientID] = receiver

		} else if req.Status == transaction.TransactionStatusApproved {
			//money is received by the client
			sender, ok := paymentCreated[req.SenderClientID]
			if !ok {
				paymentCreated[req.SenderClientID] = reports.ClientPaymentDTO{
					ClientID:                         req.SenderClientID,
					TotalPaymentSentByThisClient:     0,
					TotalPaymentReceivedByThisClient: 0,
				}
				sender = paymentCreated[req.SenderClientID]
			}
			sender.TotalPaymentSentByThisClient += req.PaymentAmount
			PaymentReport.TotalPaymentReceivedValue += req.PaymentAmount
			PaymentReport.TotalPaymentsReceived += 1
			paymentCreated[req.SenderClientID] = sender
		}
		if req.SenderClientID == clientId {
			if req.Status == "Rejected" {
				PaymentReport.RejectedPaymentRequests += 1

			} else if req.Status == transaction.TransactionStatusApproved {

				PaymentReport.ApprovedPaymentRequests += 1
			}
			PaymentReport.TotalPaymentRequests += 1
		}
	}
	if PaymentReport.TotalPaymentsSent != 0 {

		PaymentReport.AveragePaymentValue = PaymentReport.TotalPaymentSentValue / float64(PaymentReport.TotalPaymentsSent)
	}
	PaymentReport.ClientPaymentData = paymentCreated
	*report = *PaymentReport

	uow.Commit()
	return nil
}
