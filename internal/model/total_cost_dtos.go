package model

type TotalCostQuery struct {
	UserID      string `json:"user_id,omitempty" validate:"omitempty,uuid"`
	ServiceName string `json:"service_name,omitempty"`

	StartDate string `json:"start_date" validate:"required,monthyear"`
	EndDate   string `json:"end_date,omitempty" validate:"omitempty,monthyear"`
}

type TotalCostReq struct {
	UserID      *string
	ServiceName *string
	StartDate   MonthYear
	EndDate     *MonthYear
}

func (input TotalCostQuery) ToDomain() (*TotalCostReq, error) {
	var (
		userId      *string
		serviceName *string
		startDate   MonthYear
		endDate     *MonthYear
	)

	if input.UserID != "" {
		userId = &input.UserID
	}

	if input.ServiceName != "" {
		serviceName = &input.ServiceName
	}

	// --- StartDate ---
	if err := startDate.Parse(input.StartDate); err != nil {
		return nil, err
	}

	// --- EndDate ---
	if input.EndDate != "" {
		var d MonthYear
		if err := d.Parse(input.EndDate); err != nil {
			return nil, err
		}
		endDate = &d
	}

	return &TotalCostReq{
		UserID:      userId,
		ServiceName: serviceName,
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}

func (t TotalCostQuery) GetStartDate() *string {
	return &t.StartDate
}

func (t TotalCostQuery) GetEndDate() *string {
	return &t.EndDate
}
