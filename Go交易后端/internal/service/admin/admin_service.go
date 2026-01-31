package admin

// AdminService 管理端服务集合
type AdminService struct {
	User        *UserAdminService
	Currency    *CurrencyAdminService
	Wallet      *WalletAdminService
	News        *NewsAdminService
	Statistics  *StatisticsService
	System      *SystemService
	Risk        *RiskAdminService
	Transaction *TransactionAdminService
	Message     *MessageAdminService
}

// NewAdminService 创建管理端服务实例
func NewAdminService() *AdminService {
	return &AdminService{
		User:        NewUserAdminService(),
		Currency:    NewCurrencyAdminService(),
		Wallet:      NewWalletAdminService(),
		News:        NewNewsAdminService(),
		Statistics:  NewStatisticsService(),
		System:      NewSystemService(),
		Risk:        NewRiskAdminService(),
		Transaction: NewTransactionAdminService(),
		Message:     NewMessageAdminService(),
	}
}

// PageInfo 分页请求参数
type PageInfo struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

// GetOffset 计算分页偏移量
func (p *PageInfo) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	return (p.Page - 1) * p.PageSize
}

// GetLimit 获取分页限制
func (p *PageInfo) GetLimit() int {
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	return p.PageSize
}
