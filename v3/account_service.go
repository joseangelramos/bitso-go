package bitso

import (
	"context"
	"encoding/json"
	"net/http"
)

// GetAccountService get account info
type GetAccountService struct {
	c *Client
}

// Do send request
func (s *GetAccountService) Do(ctx context.Context, opts ...RequestOption) (res *AccountStatus, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/api/v3/account_status",
		secType:  secTypeSigned,
	}
	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}
	res = new(AccountStatus)
	err = json.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Account define account info
type AccountStatus struct {
	ClientId                  string `json:"client_id"`
	Status                    string `json:"status"`
	DailyLimit                string `json:"daily_limit"`
	MonthlyLimit              string `json:"monthly_limit"`
	DailyRemaining            string `json:"daily_remaining"`
	MonthlyRemaining          string `json:"monthly_remaining"`
	CellphoneNumber           string `json:"cellphone_number_stored"`
	OfficialId                string `json:"official_id"`
	ProofOfResidency          string `json:"proof_of_residency"`
	SignedContract            string `json:"signed_contract"`
	OriginOfFunds             string `json:"origin_of_funds"`
	FirstName                 string `json:"first_name"`
	LastName                  string `json:"last_name"`
	IsCellphoneNumberVerified string `json:"cellphone_number"`
	IsMailVerified            string `json:"email"`
	Email                     string `json:"email_stored"`
	ReferralCode              string `json:"referral_code"`
	CashDepositLimit          string `json:"cash_deposit_allowance"`
}
