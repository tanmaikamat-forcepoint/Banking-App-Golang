package reports

type EmployeePaymentDTO struct {
	EmpId           uint
	SalaryDisbursed float64
	MonthlySalary   float64
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
