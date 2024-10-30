package reports

type EmployeePaymentDTO struct {
	EmpId           uint
	SalaryDisbursed float64
	MonthlySalary   float64
}
type ClientPaymentDTO struct {
	ClientID                         uint
	ClientName                       string
	CreatedBy                        uint
	TotalPaymentSentByThisClient     float64
	TotalPaymentReceivedByThisClient float64
}

type SalaryReport struct {
	AverageSalary                  float64
	TotalSalaryDisbursed           float64
	ExpectedMonthlySalaryDisbursal float64
	TotalEmployees                 int
	EmployeeDisbursementData       []EmployeePaymentDTO
	StartDate                      string
	EndDate                        string
	ClientID                       uint
	ClientName                     string
}

type PaymentReport struct {
	AveragePaymentValue       float64
	TotalPaymentsSent         int
	TotalPaymentsReceived     int
	TotalPaymentReceivedValue float64
	TotalPaymentSentValue     float64
	ApprovedPaymentRequests   int
	RejectedPaymentRequests   int
	TotalPaymentRequests      int
	ClientPaymentData         map[uint]ClientPaymentDTO
	StartDate                 string
	EndDate                   string
	ClientID                  uint
	ClientName                string
}
