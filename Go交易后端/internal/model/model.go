package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

type AccountLog struct {
	ID          uint    `gorm:"primaryKey;column:id" json:"id"`
	UserID      uint    `gorm:"column:user_id" json:"user_id"`
	Value       float64 `gorm:"column:value" json:"value"`
	CreatedTime int64   `gorm:"column:created_time" json:"created_time"`
	Info        string  `gorm:"column:info" json:"info"`
	Type        int     `gorm:"column:type" json:"type"`
	CurrencyID  uint    `gorm:"column:currency" json:"currency_id"` // 统一使用CurrencyID，映射到数据库currency列
	InfoEn      string  `gorm:"column:info_en" json:"info_en"`
	InfoJp      string  `gorm:"column:info_jp" json:"info_jp"`
	InfoHk      string  `gorm:"column:info_hk" json:"info_hk"`
	InfoSpa     string  `gorm:"column:info_spa" json:"info_spa"`
	InfoKr      string  `gorm:"column:info_kr" json:"info_kr"`
	Transfered  int     `gorm:"column:transfered" json:"transfered"`
}

func (AccountLog) TableName() string {
	return "account_log"
}

type Address struct {
	ID       uint   `gorm:"primaryKey;column:id" json:"id"`
	Currency int    `gorm:"column:currency" json:"currency"`
	UserID   uint   `gorm:"column:user_id" json:"user_id"`
	Notes    string `gorm:"column:notes" json:"notes"`
	Address  string `gorm:"column:address" json:"address"`
}

func (Address) TableName() string {
	return "address"
}

type Admin struct {
	ID       uint   `gorm:"primaryKey;column:id" json:"id"`
	Username string `gorm:"column:username" json:"username"`
	Password string `gorm:"column:password" json:"password"`
	RoleID   uint   `gorm:"column:role_id" json:"role_id"`
}

func (Admin) TableName() string {
	return "admin"
}

type AdminModule struct {
	ID     uint   `gorm:"primaryKey;column:id" json:"id"`
	Name   string `gorm:"column:name" json:"name"`
	Module string `gorm:"column:module" json:"module"`
}

func (AdminModule) TableName() string {
	return "admin_module"
}

type AdminModuleAction struct {
	ID            uint   `gorm:"primaryKey;column:id" json:"id"`
	AdminModuleID uint   `gorm:"column:admin_module_id" json:"admin_module_id"`
	Name          string `gorm:"column:name" json:"name"`
	Action        string `gorm:"column:action" json:"action"`
	Level         int8   `gorm:"column:level" json:"level"`
}

func (AdminModuleAction) TableName() string {
	return "admin_module_action"
}

type AdminRole struct {
	ID      uint   `gorm:"primaryKey;column:id" json:"id"`
	Name    string `gorm:"column:name" json:"name"`
	IsSuper int8   `gorm:"column:is_super" json:"is_super"`
}

func (AdminRole) TableName() string {
	return "admin_role"
}

type AdminRolePermission struct {
	ID       uint   `gorm:"primaryKey;column:id" json:"id"`
	RoleID   uint   `gorm:"column:role_id" json:"role_id"`
	ModuleID uint   `gorm:"-" json:"module_id"`
	Module   string `gorm:"column:module" json:"module"`
	Action   string `gorm:"column:action" json:"action"`
}

func (AdminRolePermission) TableName() string {
	return "admin_role_permission"
}

type Agent struct {
	ID            uint    `gorm:"primaryKey;column:id" json:"id"`
	UserID        uint    `gorm:"column:user_id" json:"user_id"`
	Username      string  `gorm:"column:username" json:"username"`
	Password      string  `gorm:"column:password" json:"password"`
	ParentAgentID uint    `gorm:"column:parent_agent_id" json:"parent_agent_id"`
	Level         uint8   `gorm:"column:level" json:"level"`
	AgentPath     string  `gorm:"column:agent_path" json:"agent_path"`
	IsAdmin       int8    `gorm:"column:is_admin" json:"is_admin"`
	IsLock        int8    `gorm:"column:is_lock" json:"is_lock"`
	IsAddson      int8    `gorm:"column:is_addson" json:"is_addson"`
	ProLoss       float64 `gorm:"column:pro_loss" json:"pro_loss"`
	ProSer        float64 `gorm:"column:pro_ser" json:"pro_ser"`
	Status        int8    `gorm:"column:status" json:"status"`
	RegTime       int64   `gorm:"column:reg_time" json:"reg_time"`
	LockTime      int64   `gorm:"column:lock_time" json:"lock_time"`
	Money         float64 `gorm:"column:money" json:"money"`
	BtcAddress    string  `gorm:"column:btc_address" json:"btc_address"`
	UsdtAddress   string  `gorm:"column:usdt_address" json:"usdt_address"`
}

func (Agent) TableName() string {
	return "agent"
}

type AgentAdmin struct {
	ID       uint   `gorm:"primaryKey;column:id" json:"id"`
	AgentID  uint   `gorm:"column:agent_id" json:"agent_id"`
	Username string `gorm:"column:username" json:"username"`
	Password string `gorm:"column:password" json:"password"`
	RoleID   int8   `gorm:"column:role_id" json:"role_id"`
}

func (AgentAdmin) TableName() string {
	return "agent_admin"
}

type AgentLog struct {
	ID       uint    `gorm:"primaryKey;column:id" json:"id"`
	AgentID  uint    `gorm:"column:agent_id" json:"agent_id"`
	Type     uint8   `gorm:"column:type" json:"type"`
	Value    float64 `gorm:"column:value" json:"value"`
	Info     string  `gorm:"column:info" json:"info"`
	RelateID uint    `gorm:"column:relate_id" json:"relate_id"`
	AddTime  int64   `gorm:"column:add_time" json:"add_time"`
	Status   uint8   `gorm:"column:status" json:"status"`
}

func (AgentLog) TableName() string {
	return "agent_log"
}

type AgentMoneyLog struct {
	ID          uint    `gorm:"primaryKey;column:id" json:"id"`
	AgentID     uint    `gorm:"column:agent_id" json:"agent_id"`
	Type        int8    `gorm:"column:type" json:"type"`
	RelateID    uint    `gorm:"column:relate_id" json:"relate_id"`
	Before      float64 `gorm:"column:before" json:"before"`
	Change      float64 `gorm:"column:change" json:"change"`
	After       float64 `gorm:"column:after" json:"after"`
	Memo        string  `gorm:"column:memo" json:"memo"`
	CreatedTime int64   `gorm:"column:created_time" json:"created_time"`
	SonUserID   uint    `gorm:"column:son_user_id" json:"son_user_id"`
	Status      uint8   `gorm:"column:status" json:"status"`
	LegalID     uint    `gorm:"column:legal_id" json:"legal_id"`
	UpdatedTime int64   `gorm:"column:updated_time" json:"updated_time"`
}

func (AgentMoneyLog) TableName() string {
	return "agent_money_log"
}

type AgentRole struct {
	ID      uint   `gorm:"primaryKey;column:id" json:"id"`
	AgentID uint   `gorm:"column:agent_id" json:"agent_id"`
	Name    string `gorm:"column:name" json:"name"`
	IsSuper int8   `gorm:"column:is_super" json:"is_super"`
}

func (AgentRole) TableName() string {
	return "agent_role"
}

type AgentRolePermission struct {
	ID      uint   `gorm:"primaryKey;column:id" json:"id"`
	AgentID uint   `gorm:"column:agent_id" json:"agent_id"`
	RoleID  uint   `gorm:"column:role_id" json:"role_id"`
	Module  string `gorm:"column:module" json:"module"`
	Action  string `gorm:"column:action" json:"action"`
}

func (AgentRolePermission) TableName() string {
	return "agent_role_permission"
}

type Algebra struct {
	ID      uint    `gorm:"primaryKey;column:id" json:"id"`
	Name    string  `gorm:"column:name" json:"name"`
	Algebra int     `gorm:"column:algebra" json:"algebra"`
	Rate    float64 `gorm:"column:rate" json:"rate"`
}

func (Algebra) TableName() string {
	return "algebra"
}

type AreaCode struct {
	ID       uint   `gorm:"primaryKey;column:id" json:"id"`
	Name     string `gorm:"column:name" json:"name"`
	AreaCode string `gorm:"column:area_code" json:"area_code"`
}

func (AreaCode) TableName() string {
	return "area_code"
}

type AutoList struct {
	ID         uint    `gorm:"primaryKey;column:id" json:"id"`
	BuyUserID  uint    `gorm:"column:buy_user_id" json:"buy_user_id"`
	SellUserID uint    `gorm:"column:sell_user_id" json:"sell_user_id"`
	CurrencyID uint    `gorm:"column:currency_id" json:"currency_id"`
	LegalID    uint    `gorm:"column:legal_id" json:"legal_id"`
	MinPrice   float64 `gorm:"column:min_price" json:"min_price"`
	MaxPrice   float64 `gorm:"column:max_price" json:"max_price"`
	MinNumber  float64 `gorm:"column:min_number" json:"min_number"`
	MaxNumber  float64 `gorm:"column:max_number" json:"max_number"`
	NeedSecond int     `gorm:"column:need_second" json:"need_second"`
	CreateTime int64   `gorm:"column:create_time" json:"create_time"`
	IsStart    int8    `gorm:"column:is_start" json:"is_start"`
}

func (AutoList) TableName() string {
	return "auto_list"
}

type Bank struct {
	ID   uint   `gorm:"primaryKey;column:id" json:"id"`
	Name string `gorm:"column:name" json:"name"`
}

func (Bank) TableName() string {
	return "bank"
}

type C2cDeal struct {
	ID              uint    `gorm:"primaryKey;column:id" json:"id"`
	LegalDealSendID uint    `gorm:"column:legal_deal_send_id" json:"legal_deal_send_id"`
	UserID          uint    `gorm:"column:user_id" json:"user_id"`
	SellerID        uint    `gorm:"column:seller_id" json:"seller_id"`
	Number          float64 `gorm:"column:number" json:"number"`
	IsSure          int8    `gorm:"column:is_sure" json:"is_sure"`
	CreateTime      int64   `gorm:"column:create_time" json:"create_time"`
	UpdateTime      int64   `gorm:"column:update_time" json:"update_time"`
}

func (C2cDeal) TableName() string {
	return "c2c_deal"
}

type C2cDealSend struct {
	ID            uint    `gorm:"primaryKey;column:id" json:"id"`
	SellerID      uint    `gorm:"column:seller_id" json:"seller_id"`
	CurrencyID    uint    `gorm:"column:currency_id" json:"currency_id"`
	Type          string  `gorm:"column:type" json:"type"`
	Way           string  `gorm:"column:way" json:"way"`
	Price         float64 `gorm:"column:price" json:"price"`
	TotalNumber   float64 `gorm:"column:total_number" json:"total_number"`
	SurplusNumber float64 `gorm:"column:surplus_number" json:"surplus_number"`
	MinNumber     float64 `gorm:"column:min_number" json:"min_number"`
	IsDone        int8    `gorm:"column:is_done" json:"is_done"`
	CreateTime    int64   `gorm:"column:create_time" json:"create_time"`
}

func (C2cDealSend) TableName() string {
	return "c2c_deal_send"
}

type Caddress struct {
	ID         uint   `gorm:"primaryKey;column:id" json:"id"`
	TitleZh    string `gorm:"column:title_zh" json:"title_zh"`
	TitleEn    string `gorm:"column:title_en" json:"title_en"`
	TitleHk    string `gorm:"column:title_hk" json:"title_hk"`
	TitleJp    string `gorm:"column:title_jp" json:"title_jp"`
	TitleKr    string `gorm:"column:title_kr" json:"title_kr"`
	Currency   string `gorm:"column:currency" json:"currency"`
	CurrencyID uint   `gorm:"column:currency_id" json:"currency_id"`
	Address    string `gorm:"column:address" json:"address"`
	Qrcode     string `gorm:"column:qrcode" json:"qrcode"`
	ConZh      string `gorm:"column:con_zh" json:"con_zh"`
	ConEn      string `gorm:"column:con_en" json:"con_en"`
	ConHk      string `gorm:"column:con_hk" json:"con_hk"`
	ConJp      string `gorm:"column:con_jp" json:"con_jp"`
	ConKr      string `gorm:"column:con_kr" json:"con_kr"`
	Px         int    `gorm:"column:px" json:"px"`
	Status     int    `gorm:"column:status" json:"status"`
}

func (Caddress) TableName() string {
	return "caddress"
}

type CandyTransfer struct {
	ID           uint    `gorm:"primaryKey;column:id" json:"id"`
	FromUserID   uint    `gorm:"column:from_user_id" json:"from_user_id"`
	ToUserID     uint    `gorm:"column:to_user_id" json:"to_user_id"`
	TransferQty  float64 `gorm:"column:transfer_qty" json:"transfer_qty"`
	TransferRate float64 `gorm:"column:transfer_rate" json:"transfer_rate"`
	TransferFee  float64 `gorm:"column:transfer_fee" json:"transfer_fee"`
	CreateTime   int64   `gorm:"column:create_time" json:"create_time"`
}

func (CandyTransfer) TableName() string {
	return "candy_transfer"
}

type ChainHashes struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	Code      string    `gorm:"column:code" json:"code"`
	Txid      string    `gorm:"column:txid" json:"txid"`
	Amount    float64   `gorm:"column:amount" json:"amount"`
	Sender    string    `gorm:"column:sender" json:"sender"`
	Recipient string    `gorm:"column:recipient" json:"recipient"`
	Status    int8      `gorm:"column:status" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ChainHashes) TableName() string {
	return "chain_hashes"
}

type ChargeReq struct {
	ID          uint      `gorm:"primaryKey;column:id" json:"id"`
	UID         int       `gorm:"column:uid" json:"uid"`
	Amount      float64   `gorm:"column:amount" json:"amount"`
	UserAccount string    `gorm:"column:user_account" json:"user_account"`
	Status      int8      `gorm:"column:status" json:"status"`
	CurrencyID  uint      `gorm:"column:currency_id" json:"currency_id"`
	Remark      string    `gorm:"column:remark" json:"remark"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
	Image       string    `gorm:"column:image" json:"image"`
	PassmanUID  int       `gorm:"column:passman_uid" json:"passman_uid"`
	ToAddress   string    `gorm:"column:to_address" json:"to_address"`
}

func (ChargeReq) TableName() string {
	return "charge_req"
}

type ChatLog struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	Type      int8      `gorm:"column:type" json:"type"`
	Content   string    `gorm:"column:content" json:"content"`
	FromUser  uint      `gorm:"column:from_user" json:"from_user"`
	ToUser    uint      `gorm:"column:to_user" json:"to_user"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
	TradeType uint8     `gorm:"column:trade_type" json:"trade_type"`
	TradeID   uint      `gorm:"column:trade_id" json:"trade_id"`
	Readed    uint8     `gorm:"column:readed" json:"readed"`
}

func (ChatLog) TableName() string {
	return "chat_log"
}

type CoinTrade struct {
	ID          uint      `gorm:"primaryKey;column:id" json:"id"`
	UID         uint      `gorm:"column:u_id" json:"u_id"`
	CurrencyID  uint      `gorm:"column:currency_id" json:"currency_id"`
	LegalID     uint      `gorm:"column:legal_id" json:"legal_id"`
	Type        uint8     `gorm:"column:type" json:"type"`
	TargetPrice float64   `gorm:"column:target_price" json:"target_price"`
	TradePrice  float64   `gorm:"column:trade_price" json:"trade_price"`
	TradeAmount float64   `gorm:"column:trade_amount" json:"trade_amount"`
	ChargeFee   float64   `gorm:"column:charge_fee" json:"charge_fee"`
	Status      uint8     `gorm:"column:status" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (CoinTrade) TableName() string {
	return "coin_trade"
}

type Conversion struct {
	ID             uint    `gorm:"primaryKey;column:id" json:"id"`
	FormCurrencyID uint    `gorm:"column:form_currency_id" json:"form_currency_id"`
	ToCurrencyID   uint    `gorm:"column:to_currency_id" json:"to_currency_id"`
	Num            float64 `gorm:"column:num" json:"num"`
	Fee            float64 `gorm:"column:fee" json:"fee"`
	SjNum          float64 `gorm:"column:sj_num" json:"sj_num"`
	CreateTime     int64   `gorm:"column:create_time" json:"create_time"`
	UserID         uint    `gorm:"column:user_id" json:"user_id"`
}

func (Conversion) TableName() string {
	return "conversion"
}

type Currency struct {
	ID                uint    `gorm:"primaryKey;column:id" json:"id"`
	Name              string  `gorm:"column:name" json:"name"`
	GetAddress        string  `gorm:"column:get_address" json:"get_address"`
	Sort              int     `gorm:"column:sort" json:"sort"`
	Logo              string  `gorm:"column:logo" json:"logo"`
	IsDisplay         int8    `gorm:"column:is_display" json:"is_display"`
	MinNumber         float64 `gorm:"column:min_number" json:"min_number"`
	MaxNumber         float64 `gorm:"column:max_number" json:"max_number"`
	Rate              float64 `gorm:"column:rate" json:"rate"`
	IsLever           int8    `gorm:"column:is_lever" json:"is_lever"`
	IsLegal           int8    `gorm:"column:is_legal" json:"is_legal"`
	IsMatch           int8    `gorm:"column:is_match" json:"is_match"`
	IsMicro           int8    `gorm:"column:is_micro" json:"is_micro"`
	Insurancable      uint8   `gorm:"column:insurancable" json:"insurancable"`
	ShowLegal         int8    `gorm:"column:show_legal" json:"show_legal"`
	Type              string  `gorm:"column:type" json:"type"`
	BlackLimt         int     `gorm:"column:black_limt" json:"black_limt"`
	Key               string  `gorm:"column:key" json:"key"`
	ContractAddress   string  `gorm:"column:contract_address" json:"contract_address"`
	TotalAccount      string  `gorm:"column:total_account" json:"total_account"`
	CollectAccount    string  `gorm:"column:collect_account" json:"collect_account"`
	CurrencyDecimals  int     `gorm:"column:currency_decimals" json:"currency_decimals"`
	RmbRelation       float64 `gorm:"column:rmb_relation" json:"rmb_relation"`
	DecimalScale      int     `gorm:"column:decimal_scale" json:"decimal_scale"`
	ChainFee          float64 `gorm:"column:chain_fee" json:"chain_fee"`
	Price             float64 `gorm:"column:price" json:"price"`
	MicroTradeFee     float64 `gorm:"column:micro_trade_fee" json:"micro_trade_fee"`
	MicroMin          float64 `gorm:"column:micro_min" json:"micro_min"`
	MicroMax          float64 `gorm:"column:micro_max" json:"micro_max"`
	MicroHoldtradeMax int     `gorm:"column:micro_holdtrade_max" json:"micro_holdtrade_max"`
	IsChange          int8    `gorm:"column:is_change" json:"is_change"`
	UpdateTime        int64   `gorm:"column:update_time" json:"update_time"`
	CreateTime        int64   `gorm:"column:create_time" json:"create_time"`

	// 风控相关字段
	RiskProbEnabled       int8    `gorm:"column:risk_prob_enabled;default:0" json:"risk_prob_enabled"`
	RiskProfitProbability int     `gorm:"column:risk_profit_probability;default:50" json:"risk_profit_probability"`
	RiskMoneyEnabled      int8    `gorm:"column:risk_money_enabled;default:0" json:"risk_money_enabled"`
	RiskMoneyMin          float64 `gorm:"column:risk_money_min;default:0.0" json:"risk_money_min"`
	RiskMoneyMax          float64 `gorm:"column:risk_money_max;default:0.0" json:"risk_money_max"`
	RiskMoneyResult       int8    `gorm:"column:risk_money_result;default:0" json:"risk_money_result"`
	RiskTimeEnabled       int8    `gorm:"column:risk_time_enabled;default:0" json:"risk_time_enabled"`
	RiskTimeStart         string  `gorm:"column:risk_time_start;default:''" json:"risk_time_start"`
	RiskTimeEnd           string  `gorm:"column:risk_time_end;default:''" json:"risk_time_end"`
	RiskTimeResult        int8    `gorm:"column:risk_time_result;default:0" json:"risk_time_result"`
}

func (Currency) TableName() string {
	return "currency"
}

type CurrencyMatch struct {
	ID              uint    `gorm:"primaryKey;column:id" json:"id"`
	LegalID         uint    `gorm:"column:legal_id" json:"legal_id"`
	CurrencyID      uint    `gorm:"column:currency_id" json:"currency_id"`
	IsDisplay       int8    `gorm:"column:is_display" json:"is_display"`
	MarketFrom      int8    `gorm:"column:market_from" json:"market_from"`
	OpenTransaction int8    `gorm:"column:open_transaction" json:"open_transaction"`
	OpenLever       int8    `gorm:"column:open_lever" json:"open_lever"`
	OpenMicrotrade  int8    `gorm:"column:open_microtrade" json:"open_microtrade"`
	Sort            int     `gorm:"column:sort" json:"sort"`
	MicroTradeFee   float64 `gorm:"column:micro_trade_fee" json:"micro_trade_fee"`
	LeverShareNum   float64 `gorm:"column:lever_share_num" json:"lever_share_num"`
	Spread          float64 `gorm:"column:spread" json:"spread"`
	Overnight       float64 `gorm:"column:overnight" json:"overnight"`
	LeverTradeFee   float64 `gorm:"column:lever_trade_fee" json:"lever_trade_fee"`
	LeverMinShare   uint    `gorm:"column:lever_min_share" json:"lever_min_share"`
	LeverMaxShare   uint    `gorm:"column:lever_max_share" json:"lever_max_share"`
	FluctuateMin    float64 `gorm:"column:fluctuate_min" json:"fluctuate_min"`
	FluctuateMax    float64 `gorm:"column:fluctuate_max" json:"fluctuate_max"`
	RiskGroupResult int8    `gorm:"column:risk_group_result" json:"risk_group_result"`
	CreateTime      int64   `gorm:"column:create_time" json:"create_time"`
}

func (CurrencyMatch) TableName() string {
	return "currency_matches"
}

type CurrencyQuotation struct {
	ID         uint    `gorm:"primaryKey;column:id" json:"id"`
	MatchID    uint    `gorm:"column:match_id" json:"match_id"`
	LegalID    uint    `gorm:"column:legal_id" json:"legal_id"`
	CurrencyID uint    `gorm:"column:currency_id" json:"currency_id"`
	Change     string  `gorm:"column:change" json:"change"`
	Volume     float64 `gorm:"column:volume" json:"volume"`
	Price      float64 `gorm:"column:now_price" json:"price"`
	NowPrice   float64 `gorm:"column:now_price" json:"now_price"`
	AddTime    int64   `gorm:"column:add_time" json:"add_time"`
}

func (CurrencyQuotation) TableName() string {
	return "currency_quotation"
}

type FailedJob struct {
	ID         uint64    `gorm:"primaryKey;column:id" json:"id"`
	Connection string    `gorm:"column:connection" json:"connection"`
	Queue      string    `gorm:"column:queue" json:"queue"`
	Payload    string    `gorm:"column:payload" json:"payload"`
	Exception  string    `gorm:"column:exception" json:"exception"`
	FailedAt   time.Time `gorm:"column:failed_at" json:"failed_at"`
}

func (FailedJob) TableName() string {
	return "failed_jobs"
}

type FalseData struct {
	ID      uint   `gorm:"primaryKey;column:id" json:"id"`
	Address string `gorm:"column:address" json:"address"`
	Price   string `gorm:"column:price" json:"price"`
	Time    int64  `gorm:"column:time" json:"time"`
}

func (FalseData) TableName() string {
	return "false_data"
}

type Feedback struct {
	ID           uint   `gorm:"primaryKey;column:id" json:"id"`
	UserID       uint   `gorm:"column:user_id" json:"user_id"`
	Content      string `gorm:"column:content" json:"content"`
	Img          string `gorm:"column:img" json:"img"`
	CreateTime   int64  `gorm:"column:create_time" json:"create_time"`
	ReplyTime    int64  `gorm:"column:reply_time" json:"reply_time"`
	ReplyContent string `gorm:"column:reply_content" json:"reply_content"`
	IsReply      int8   `gorm:"column:is_reply" json:"is_reply"`
}

func (Feedback) TableName() string {
	return "feedback"
}

type FeedBack = Feedback

type FlashAgainst struct {
	ID               uint    `gorm:"primaryKey;column:id" json:"id"`
	UserID           uint    `gorm:"column:user_id" json:"user_id"`
	LeftCurrencyID   uint    `gorm:"column:left_currency_id" json:"left_currency_id"`
	RightCurrencyID  uint    `gorm:"column:right_currency_id" json:"right_currency_id"`
	Num              float64 `gorm:"column:num" json:"num"`
	Fee              float64 `gorm:"column:fee" json:"fee"`
	AbsoluteQuantity float64 `gorm:"column:absolute_quantity" json:"absolute_quantity"`
	MarketPrice      float64 `gorm:"column:market_price" json:"market_price"`
	Price            float64 `gorm:"column:price" json:"price"`
	Status           int8    `gorm:"column:status" json:"status"`
	ReviewTime       int64   `gorm:"column:review_time" json:"review_time"`
	CreateTime       int64   `gorm:"column:create_time" json:"create_time"`
}

func (FlashAgainst) TableName() string {
	return "flash_against"
}

type HistoricalData struct {
	ID        uint   `gorm:"primaryKey;column:id" json:"id"`
	Type      string `gorm:"column:type" json:"type"`
	StartTime int64  `gorm:"column:start_time" json:"start_time"`
	EndTime   int64  `gorm:"column:end_time" json:"end_time"`
	Data      string `gorm:"column:data" json:"data"`
}

func (HistoricalData) TableName() string {
	return "historical_data"
}

type HuobiSymbols struct {
	ID              uint   `gorm:"primaryKey;column:id" json:"id"`
	BaseCurrency    string `gorm:"column:base-currency" json:"base-currency"`
	QuoteCurrency   string `gorm:"column:quote-currency" json:"quote-currency"`
	PricePrecision  int    `gorm:"column:price-precision" json:"price-precision"`
	AmountPrecision int    `gorm:"column:amount-precision" json:"amount-precision"`
	SymbolPartition string `gorm:"column:symbol-partition" json:"symbol-partition"`
	Symbol          string `gorm:"column:symbol" json:"symbol"`
}

func (HuobiSymbols) TableName() string {
	return "huobi_symbols"
}

type InsuranceClaimApplies struct {
	ID              uint      `gorm:"primaryKey;column:id" json:"id"`
	UserID          uint      `gorm:"column:user_id" json:"user_id"`
	UserInsuranceID uint      `gorm:"column:user_insurance_id" json:"user_insurance_id"`
	InsuranceType   uint8     `gorm:"column:insurance_type" json:"insurance_type"`
	ApplyStatus     uint8     `gorm:"column:apply_status" json:"apply_status"`
	Compensate      float64   `gorm:"column:compensate" json:"compensate"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`
	Operator        string    `gorm:"column:operator" json:"operator"`
	RefuseReason    string    `gorm:"column:refuse_reason" json:"refuse_reason"`
}

func (InsuranceClaimApplies) TableName() string {
	return "insurance_claim_applies"
}

type InsuranceRules struct {
	ID              uint    `gorm:"primaryKey;column:id" json:"id"`
	InsuranceTypeID uint16  `gorm:"column:insurance_type_id" json:"insurance_type_id"`
	Amount          float64 `gorm:"column:amount" json:"amount"`
	PlaceAnOrderMax float64 `gorm:"column:place_an_order_max" json:"place_an_order_max"`
	ExistingNumber  uint16  `gorm:"column:existing_number" json:"existing_number"`
}

func (InsuranceRules) TableName() string {
	return "insurance_rules"
}

type InsuranceTypes struct {
	ID                         uint    `gorm:"primaryKey;column:id" json:"id"`
	Name                       string  `gorm:"column:name" json:"name"`
	CurrencyID                 uint16  `gorm:"column:currency_id" json:"currency_id"`
	Type                       uint8   `gorm:"column:type" json:"type"`
	MinAmount                  float64 `gorm:"column:min_amount" json:"min_amount"`
	MaxAmount                  float64 `gorm:"column:max_amount" json:"max_amount"`
	InsuranceAssets            float64 `gorm:"column:insurance_assets" json:"insurance_assets"`
	ProfitTerminationCondition float64 `gorm:"column:profit_termination_condition" json:"profit_termination_condition"`
	DefectiveClaimsCondition   float64 `gorm:"column:defective_claims_condition" json:"defective_claims_condition"`
	DefectiveClaimsCondition2  float64 `gorm:"column:defective_claims_condition2" json:"defective_claims_condition2"`
	ClaimsTimesDaily           uint8   `gorm:"column:claims_times_daily" json:"claims_times_daily"`
	AutoClaim                  uint8   `gorm:"column:auto_claim" json:"auto_claim"`
	ClaimRate                  float64 `gorm:"column:claim_rate" json:"claim_rate"`
	ClaimDirection             uint8   `gorm:"column:claim_direction" json:"claim_direction"`
	Status                     uint8   `gorm:"column:status" json:"status"`
	IsTAdd1                    uint8   `gorm:"column:is_t_add_1" json:"is_t_add_1"`
}

func (InsuranceTypes) TableName() string {
	return "insurance_types"
}

type IPWhitelist struct {
	ID      uint      `gorm:"primaryKey;column:id" json:"id"`
	Ipadder string    `gorm:"column:ipadder" json:"ipadder"`
	Type    int8      `gorm:"column:type" json:"type"`
	Status  int8      `gorm:"column:status" json:"status"`
	Ctime   time.Time `gorm:"column:ctime" json:"ctime"`
	Utime   time.Time `gorm:"column:utime" json:"utime"`
}

func (IPWhitelist) TableName() string {
	return "ip_whiltlist"
}

type Job struct {
	ID          uint64 `gorm:"primaryKey;column:id" json:"id"`
	Queue       string `gorm:"column:queue" json:"queue"`
	Payload     string `gorm:"column:payload" json:"payload"`
	Attempts    uint8  `gorm:"column:attempts" json:"attempts"`
	ReservedAt  uint   `gorm:"column:reserved_at" json:"reserved_at"`
	AvailableAt uint   `gorm:"column:available_at" json:"available_at"`
	CreatedAt   uint   `gorm:"column:created_at" json:"created_at"`
}

func (Job) TableName() string {
	return "jobs"
}

type Lab struct {
	ID          uint    `gorm:"primaryKey;column:id" json:"id"`
	CID         uint    `gorm:"column:c_id" json:"c_id"`
	CName       string  `gorm:"column:c_name" json:"c_name"`
	Lang        string  `gorm:"column:lang" json:"lang"`
	Title       string  `gorm:"column:title" json:"title"`
	Recommend   int8    `gorm:"column:recommend" json:"recommend"`
	Audit       int8    `gorm:"column:audit" json:"audit"`
	Display     int8    `gorm:"column:display" json:"display"`
	Discuss     int8    `gorm:"column:discuss" json:"discuss"`
	Author      string  `gorm:"column:author" json:"author"`
	BrowseGrant int8    `gorm:"column:browse_grant" json:"browse_grant"`
	Keyword     string  `gorm:"column:keyword" json:"keyword"`
	Abstract    string  `gorm:"column:abstract" json:"abstract"`
	Content     string  `gorm:"column:content" json:"content"`
	Views       int     `gorm:"column:views" json:"views"`
	CreateTime  int64   `gorm:"column:create_time" json:"create_time"`
	UpdateTime  int64   `gorm:"column:update_time" json:"update_time"`
	Thumbnail   string  `gorm:"column:thumbnail" json:"thumbnail"`
	Cover       string  `gorm:"column:cover" json:"cover"`
	Sorts       int     `gorm:"column:sorts" json:"sorts"`
	StartTime   int64   `gorm:"column:start_time" json:"start_time"`
	EndTime     int64   `gorm:"column:end_time" json:"end_time"`
	Total       float64 `gorm:"column:total" json:"total"`
	Buymoney    float64 `gorm:"column:buymoney" json:"buymoney"`
	Delay       int     `gorm:"column:delay" json:"delay"`
	Sczq        int     `gorm:"column:sczq" json:"sczq"`
}

func (Lab) TableName() string {
	return "lab"
}

type LabList struct {
	ID           uint    `gorm:"primaryKey;column:id" json:"id"`
	UserID       uint    `gorm:"column:user_id" json:"user_id"`
	Amount       float64 `gorm:"column:amount" json:"amount"`
	Status       int8    `gorm:"column:status" json:"status"`
	CurrencyID   uint    `gorm:"column:currency_id" json:"currency_id"`
	CurrencyName string  `gorm:"column:currency_name" json:"currency_name"`
	CreatedAt    int     `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    int     `gorm:"column:updated_at" json:"updated_at"`
	EndtimeAt    int     `gorm:"column:endtime_at" json:"endtime_at"`
	PayWalletID  uint    `gorm:"column:pay_wallet_id" json:"pay_wallet_id"`
	Price        float64 `gorm:"column:price" json:"price"`
	PriceSell    float64 `gorm:"column:price_sell" json:"price_sell"`
	LabID        uint    `gorm:"column:lab_id" json:"lab_id"`
}

func (LabList) TableName() string {
	return "lab_lists"
}

type LbxHashes struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	WalletID  uint      `gorm:"column:wallet_id" json:"wallet_id"`
	Txid      string    `gorm:"column:txid" json:"txid"`
	Type      int8      `gorm:"column:type" json:"type"`
	Amount    float64   `gorm:"column:amount" json:"amount"`
	Status    int8      `gorm:"column:status" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (LbxHashes) TableName() string {
	return "lbx_hashes"
}

type LegalDeal struct {
	ID              uint    `gorm:"primaryKey;column:id" json:"id"`
	LegalDealSendID uint    `gorm:"column:legal_deal_send_id" json:"legal_deal_send_id"`
	UserID          uint    `gorm:"column:user_id" json:"user_id"`
	SellerID        uint    `gorm:"column:seller_id" json:"seller_id"`
	Number          float64 `gorm:"column:number" json:"number"`
	IsSure          int8    `gorm:"column:is_sure" json:"is_sure"`
	PayOrdersImg    string  `gorm:"column:pay_orders_img" json:"pay_orders_img"`
	CreateTime      int64   `gorm:"column:create_time" json:"create_time"`
	UpdateTime      int64   `gorm:"column:update_time" json:"update_time"`
}

func (LegalDeal) TableName() string {
	return "legal_deal"
}

type LegalDealSend struct {
	ID            uint    `gorm:"primaryKey;column:id" json:"id"`
	SellerID      uint    `gorm:"column:seller_id" json:"seller_id"`
	CurrencyID    uint    `gorm:"column:currency_id" json:"currency_id"`
	Type          string  `gorm:"column:type" json:"type"`
	Way           string  `gorm:"column:way" json:"way"`
	Price         float64 `gorm:"column:price" json:"price"`
	TotalNumber   float64 `gorm:"column:total_number" json:"total_number"`
	SurplusNumber float64 `gorm:"column:surplus_number" json:"surplus_number"`
	MinNumber     float64 `gorm:"column:min_number" json:"min_number"`
	MaxNumber     float64 `gorm:"column:max_number" json:"max_number"`
	IsDone        int8    `gorm:"column:is_done" json:"is_done"`
	CreateTime    int64   `gorm:"column:create_time" json:"create_time"`
	IsShelves     int     `gorm:"column:is_shelves" json:"is_shelves"`
	IsSendback    int     `gorm:"column:is_sendback" json:"is_sendback"`
}

func (LegalDealSend) TableName() string {
	return "legal_deal_send"
}

type LegalOrder struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	UserID     uint      `gorm:"column:user_id" json:"user_id"`
	Status     int       `gorm:"column:status" json:"status"`
	CreatedAt  uint      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
	Amount     float64   `gorm:"column:amount" json:"amount"`
	UsdtAmount float64   `gorm:"column:usdt_amount" json:"usdt_amount"`
	Rate       float64   `gorm:"column:rate" json:"rate"`
	PayWay     string    `gorm:"column:pay_way" json:"pay_way"`
	PayAccount string    `gorm:"column:pay_account" json:"pay_account"`
	PayTime    time.Time `gorm:"column:pay_time" json:"pay_time"`
	URL        string    `gorm:"column:url" json:"url"`
	Type       string    `gorm:"column:type" json:"type"`
	StoreID    uint      `gorm:"column:store_id" json:"store_id"`
}

func (LegalOrder) TableName() string {
	return "legal_order"
}

type LegalStore struct {
	ID            uint    `gorm:"primaryKey;column:id" json:"id"`
	Name          string  `gorm:"column:name" json:"name"`
	BankName      string  `gorm:"column:bank_name" json:"bank_name"`
	BankAccount   string  `gorm:"column:bank_account" json:"bank_account"`
	BankUser      string  `gorm:"column:bank_user" json:"bank_user"`
	BankSubname   string  `gorm:"column:bank_subname" json:"bank_subname"`
	AlipayAccount string  `gorm:"column:alipay_account" json:"alipay_account"`
	AlipayQrcode  string  `gorm:"column:alipay_qrcode" json:"alipay_qrcode"`
	WechatAccount string  `gorm:"column:wechat_account" json:"wechat_account"`
	WechatQrcode  string  `gorm:"column:wechat_qrcode" json:"wechat_qrcode"`
	MinNum        float64 `gorm:"column:min_num" json:"min_num"`
	MaxNum        float64 `gorm:"column:max_num" json:"max_num"`
	MinNumWid     float64 `gorm:"column:min_num_wid" json:"min_num_wid"`
	MaxNumWid     float64 `gorm:"column:max_num_wid" json:"max_num_wid"`
	Rate          float64 `gorm:"column:rate" json:"rate"`
	RateSell      float64 `gorm:"column:rate_sell" json:"rate_sell"`
	UpdatedAt     uint    `gorm:"column:updated_at" json:"updated_at"`
	CreatedAt     uint    `gorm:"column:created_at" json:"created_at"`
}

func (LegalStore) TableName() string {
	return "legal_store"
}

type Level struct {
	ID               uint    `gorm:"primaryKey;column:id" json:"id"`
	Name             string  `gorm:"column:name" json:"name"`
	FillCurrency     float64 `gorm:"column:fill_currency" json:"fill_currency"`
	DirectDriveCount int     `gorm:"column:direct_drive_count" json:"direct_drive_count"`
	DirectDrivePrice float64 `gorm:"column:direct_drive_price" json:"direct_drive_price"`
	MaxAlgebra       int     `gorm:"column:max_algebra" json:"max_algebra"`
	Level            int     `gorm:"column:level" json:"level"`
}

func (Level) TableName() string {
	return "level"
}

type LeverMultiple struct {
	ID         uint   `gorm:"primaryKey;column:id" json:"id"`
	Type       int    `gorm:"column:type" json:"type"`                                      // 1=倍数, 2=手数
	Value      string `gorm:"column:value" json:"value"`                                    // 杠杆值
	CurrencyID *uint  `gorm:"column:currency_id;default:null" json:"currency_id,omitempty"` // 可选，null表示全局配置
}

func (LeverMultiple) TableName() string {
	return "lever_multiple"
}

type LeverToLegal struct {
	ID      uint    `gorm:"primaryKey;column:id" json:"id"`
	UserID  uint    `gorm:"column:user_id" json:"user_id"`
	Number  float64 `gorm:"column:number" json:"number"`
	AddTime int64   `gorm:"column:add_time" json:"add_time"`
	Type    int     `gorm:"column:type" json:"type"`
	Status  int     `gorm:"column:status" json:"status"`
}

func (LeverToLegal) TableName() string {
	return "lever_tolegal"
}

// LeverTransaction 合约交易记录模型
type LeverTransaction struct {
	ID                 uint     `gorm:"primaryKey;column:id" json:"id"`
	Type               int8     `gorm:"column:type" json:"type"`                       // 买卖类型:1=买入(做多),2=卖出(做空)
	OrderType          int8     `gorm:"column:order_type;default:1" json:"order_type"` // 订单类型:1=市价单,2=限价单
	UserID             uint     `gorm:"column:user_id" json:"user_id"`
	CurrencyID         uint     `gorm:"column:currency" json:"currency_id"`
	LegalID            uint     `gorm:"column:legal" json:"legal_id"`
	Currency           uint     `gorm:"column:currency" json:"currency"`
	Legal              uint     `gorm:"column:legal" json:"legal"`
	OriginPrice        float64  `gorm:"column:origin_price" json:"origin_price"`                       // 原始价格
	LimitPrice         *float64 `gorm:"column:limit_price" json:"limit_price,omitempty"`               // 限价价格(限价单时使用)
	Price              float64  `gorm:"column:price" json:"price"`                                     // 开仓价格(点差处理之后)
	UpdatePrice        float64  `gorm:"column:update_price" json:"update_price"`                       // 当前价格
	TargetProfitPrice  float64  `gorm:"column:target_profit_price" json:"target_profit_price"`         // 止盈价格(旧字段)
	StopLossPrice      float64  `gorm:"column:stop_loss_price" json:"stop_loss_price"`                 // 止损价格(旧字段)
	TakeProfitAmount   *float64 `gorm:"column:take_profit_amount" json:"take_profit_amount,omitempty"` // 止盈金额(USDT绝对值)
	StopLossAmount     *float64 `gorm:"column:stop_loss_amount" json:"stop_loss_amount,omitempty"`     // 止损金额(USDT绝对值)
	Share              uint     `gorm:"column:share" json:"share"`                                     // 手数
	Number             float64  `gorm:"column:number" json:"number"`                                   // 数量
	Multiple           int      `gorm:"column:multiple" json:"multiple"`                               // 杠杆倍数
	OriginCautionMoney float64  `gorm:"column:origin_caution_money" json:"origin_caution_money"`       // 初始保证金
	CautionMoney       float64  `gorm:"column:caution_money" json:"caution_money"`                     // 当前可用保证金
	ActualMargin       *float64 `gorm:"column:actual_margin" json:"actual_margin,omitempty"`           // 实际保证金(扣除手续费后)
	FactProfits        float64  `gorm:"column:fact_profits" json:"fact_profits"`                       // 最终盈亏
	TradeFee           float64  `gorm:"column:trade_fee" json:"trade_fee"`                             // 交易手续费
	Overnight          float64  `gorm:"column:overnight" json:"overnight"`                             // 隔夜费率
	OvernightMoney     float64  `gorm:"column:overnight_money" json:"overnight_money"`                 // 隔夜费金额
	Status             int8     `gorm:"column:status" json:"status"`                                   // 状态:0=挂单中,1=交易中,2=平仓中,3=已平仓,4=已撤单
	CloseType          *int8    `gorm:"column:close_type" json:"close_type,omitempty"`                 // 平仓类型:1=手动,2=爆仓,3=止盈,4=止损
	ClosePrice         *float64 `gorm:"column:close_price" json:"close_price,omitempty"`               // 平仓价格
	Settled            int8     `gorm:"column:settled" json:"settled"`                                 // 结算状态:0=未结算,1=已结算
	CreateTime         int64    `gorm:"column:create_time" json:"create_time"`                         // 下单时间
	TransactionTime    float64  `gorm:"column:transaction_time" json:"transaction_time"`               // 成交时间
	UpdateTime         float64  `gorm:"column:update_time" json:"update_time"`                         // 价格刷新时间
	HandleTime         float64  `gorm:"column:handle_time" json:"handle_time"`                         // 平仓时间
	CompleteTime       float64  `gorm:"column:complete_time" json:"complete_time"`                     // 完成时间
	AgentPath          string   `gorm:"column:agent_path" json:"agent_path"`                           // 代理商关系
	Locked             int      `gorm:"column:locked" json:"locked"`                                   // 是否锁仓
}

func (LeverTransaction) TableName() string {
	return "lever_transaction"
}

type Market struct {
	ID                uint   `gorm:"primaryKey;column:id" json:"id"`
	Name              string `gorm:"column:name" json:"name"`
	Symbol            string `gorm:"column:symbol" json:"symbol"`
	WebsiteSlug       string `gorm:"column:website_slug" json:"website_slug"`
	Rank              int    `gorm:"column:rank" json:"rank"`
	CirculatingSupply uint64 `gorm:"column:circulating_supply" json:"circulating_supply"`
	TotalSupply       uint64 `gorm:"column:total_supply" json:"total_supply"`
	MaxSupply         uint64 `gorm:"column:max_supply" json:"max_supply"`
	Quotes            string `gorm:"column:quotes" json:"quotes"`
	LastUpdated       uint   `gorm:"column:last_updated" json:"last_updated"`
}

func (Market) TableName() string {
	return "market"
}

type MarketDay struct {
	ID         uint    `gorm:"primaryKey;column:id" json:"id"`
	CurrencyID uint    `gorm:"column:currency_id" json:"currency_id"`
	LegalID    uint    `gorm:"column:legal_id" json:"legal_id"`
	StartPrice float64 `gorm:"column:start_price" json:"start_price"`
	EndPrice   float64 `gorm:"column:end_price" json:"end_price"`
	Highest    float64 `gorm:"column:highest" json:"highest"`
	Mminimum   float64 `gorm:"column:mminimum" json:"mminimum"`
	Number     float64 `gorm:"column:number" json:"number"`
	Times      string  `gorm:"column:times" json:"times"`
	MarID      string  `gorm:"column:mar_id" json:"mar_id"`
	Type       int     `gorm:"column:type" json:"type"`
}

func (MarketDay) TableName() string {
	return "market_day"
}

type MarketHour struct {
	ID         uint    `gorm:"primaryKey;column:id" json:"id"`
	CurrencyID uint    `gorm:"column:currency_id" json:"currency_id"`
	LegalID    uint    `gorm:"column:legal_id" json:"legal_id"`
	Open       float64 `gorm:"column:start_price" json:"open"`
	Close      float64 `gorm:"column:end_price" json:"close"`
	High       float64 `gorm:"column:highest" json:"high"`
	Low        float64 `gorm:"column:mminimum" json:"low"`
	Timestamp  int64   `gorm:"column:day_time" json:"timestamp"`
	Type       int8    `gorm:"column:type" json:"type"`
	Volume     float64 `gorm:"column:number" json:"volume"`
	SignTime   string  `gorm:"column:mar_id" json:"sign_time"`
	Period     string  `gorm:"column:period" json:"period"`
	Sign       int8    `gorm:"column:sign" json:"sign"`
}

func (MarketHour) TableName() string {
	return "market_hour"
}

type MicroNumber struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	CurrencyID uint      `gorm:"column:currency_id" json:"currency_id"`
	Number     float64   `gorm:"column:number" json:"number"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (MicroNumber) TableName() string {
	return "micro_numbers"
}

// MicroOrder 秒合约订单
// 状态定义: Status: 0=进行中, 1=已结算
// 方向定义: Type: 1=买涨(rise), 2=买跌(fall)
// 结果定义: ProfitResult: 1=盈利, -1=亏损 (无平局)
type MicroOrder struct {
	ID              uint      `gorm:"primaryKey;column:id" json:"id"`
	UserID          uint      `gorm:"column:user_id" json:"user_id"`
	MatchID         uint      `gorm:"column:match_id" json:"match_id"`                   // [废弃] 保留字段兼容旧数据
	CurrencyID      uint      `gorm:"column:currency_id" json:"currency_id"`             // 交易币种ID
	Type            int8      `gorm:"column:type" json:"type"`                           // 方向: 1=买涨, 2=买跌
	IsInsurance     int8      `gorm:"column:is_insurance" json:"is_insurance"`           // [废弃] 保留字段兼容旧数据
	Seconds         uint      `gorm:"column:seconds" json:"seconds"`                     // 周期秒数: 30/60/120/180/300
	Number          float64   `gorm:"column:number" json:"number"`                       // 投入金额(USDT)
	OpenPrice       float64   `gorm:"column:open_price" json:"open_price"`               // 开仓价格
	EndPrice        float64   `gorm:"column:end_price" json:"end_price"`                 // 结算价格
	Fee             float64   `gorm:"column:fee" json:"fee"`                             // [废弃] 手续费,固定为0
	ProfitRatio     float64   `gorm:"column:profit_ratio" json:"profit_ratio"`           // 盈利率(如0.85表示85%)
	FactProfits     float64   `gorm:"column:fact_profits" json:"fact_profits"`           // 实际盈亏金额
	Status          int8      `gorm:"column:status" json:"status"`                       // 状态: 0=进行中, 1=已结算
	PreProfitResult int8      `gorm:"column:pre_profit_result" json:"pre_profit_result"` // 风控预设结果: 1=盈, -1=亏, 0=无预设
	ProfitResult    int8      `gorm:"column:profit_result" json:"profit_result"`         // 最终结果: 1=盈利, -1=亏损
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`               // 创建时间(下单时间)
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`               // 更新时间(结算时间)
	HandledAt       time.Time `gorm:"column:handled_at" json:"handled_at"`               // [废弃] 保留字段兼容旧数据
	CompleteAt      time.Time `gorm:"column:complete_at" json:"complete_at"`             // [废弃] 保留字段兼容旧数据
	ReturnAt        time.Time `gorm:"column:return_at" json:"return_at"`                 // [废弃] 保留字段兼容旧数据
	AgentPath       string    `gorm:"column:agent_path" json:"agent_path"`               // [废弃] 保留字段兼容旧数据
}

func (MicroOrder) TableName() string {
	return "micro_orders"
}

type MicroOrderCopy struct {
	ID              uint      `gorm:"primaryKey;column:id" json:"id"`
	UserID          uint      `gorm:"column:user_id" json:"user_id"`
	MatchID         uint      `gorm:"column:match_id" json:"match_id"`
	CurrencyID      uint      `gorm:"column:currency_id" json:"currency_id"`
	Type            int8      `gorm:"column:type" json:"type"`
	Seconds         uint      `gorm:"column:seconds" json:"seconds"`
	Number          float64   `gorm:"column:number" json:"number"`
	OpenPrice       float64   `gorm:"column:open_price" json:"open_price"`
	EndPrice        float64   `gorm:"column:end_price" json:"end_price"`
	Fee             float64   `gorm:"column:fee" json:"fee"`
	ProfitRatio     float64   `gorm:"column:profit_ratio" json:"profit_ratio"`
	FactProfits     float64   `gorm:"column:fact_profits" json:"fact_profits"`
	Status          int8      `gorm:"column:status" json:"status"`
	PreProfitResult int8      `gorm:"column:pre_profit_result" json:"pre_profit_result"`
	ProfitResult    int8      `gorm:"column:profit_result" json:"profit_result"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`
	HandledAt       time.Time `gorm:"column:handled_at" json:"handled_at"`
	CompleteAt      time.Time `gorm:"column:complete_at" json:"complete_at"`
}

func (MicroOrderCopy) TableName() string {
	return "micro_orders_copy"
}

// MicroSeconds 秒合约周期配置表
// 固定5档周期: 30/60/120/180/300秒
// 盈利率可后台配置，周期不可修改
type MicroSeconds struct {
	ID          uint      `gorm:"primaryKey;column:id" json:"id"`
	Seconds     uint      `gorm:"column:seconds" json:"seconds"`           // 周期秒数: 30/60/120/180/300
	Status      int8      `gorm:"column:status" json:"status"`             // 状态: 1=启用, 0=禁用
	ProfitRatio float64   `gorm:"column:profit_ratio" json:"profit_ratio"` // 盈利率(如 0.85 表示 85%)
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (MicroSeconds) TableName() string {
	return "micro_seconds"
}

type Migration struct {
	ID        uint   `gorm:"primaryKey;column:id" json:"id"`
	Migration string `gorm:"column:migration" json:"migration"`
	Batch     int    `gorm:"column:batch" json:"batch"`
}

func (Migration) TableName() string {
	return "migrations"
}

type Myquotation struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	Open      float64   `gorm:"column:open" json:"open"`
	High      float64   `gorm:"column:high" json:"high"`
	Low       float64   `gorm:"column:low" json:"low"`
	Close     float64   `gorm:"column:close" json:"close"`
	Symbol    string    `gorm:"column:symbol" json:"symbol"`
	Base      string    `gorm:"column:base" json:"base"`
	Target    string    `gorm:"column:target" json:"target"`
	Itime     int64     `gorm:"column:itime" json:"itime"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	Vol       float64   `gorm:"column:vol" json:"vol"`
}

func (Myquotation) TableName() string {
	return "myquotation"
}

type Needle struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	Open      float64   `gorm:"column:open" json:"open"`
	High      float64   `gorm:"column:high" json:"high"`
	Low       float64   `gorm:"column:low" json:"low"`
	Close     float64   `gorm:"column:close" json:"close"`
	Symbol    string    `gorm:"column:symbol" json:"symbol"`
	Base      string    `gorm:"column:base" json:"base"`
	Target    string    `gorm:"column:target" json:"target"`
	Itime     int64     `gorm:"column:itime" json:"itime"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (Needle) TableName() string {
	return "needle"
}

type News struct {
	ID          uint   `gorm:"primaryKey;column:id" json:"id"`
	CID         uint   `gorm:"column:c_id" json:"c_id"`
	Lang        string `gorm:"column:lang" json:"lang"`
	Title       string `gorm:"column:title" json:"title"`
	Recommend   int8   `gorm:"column:recommend" json:"recommend"`
	Audit       int8   `gorm:"column:audit" json:"audit"`
	Display     int8   `gorm:"column:display" json:"display"`
	Discuss     int8   `gorm:"column:discuss" json:"discuss"`
	Author      string `gorm:"column:author" json:"author"`
	BrowseGrant int8   `gorm:"column:browse_grant" json:"browse_grant"`
	Keyword     string `gorm:"column:keyword" json:"keyword"`
	Abstract    string `gorm:"column:abstract" json:"abstract"`
	Content     string `gorm:"column:content" json:"content"`
	Views       int    `gorm:"column:views" json:"views"`
	CreateTime  int64  `gorm:"column:create_time" json:"create_time"`
	UpdateTime  int64  `gorm:"column:update_time" json:"update_time"`
	Thumbnail   string `gorm:"column:thumbnail" json:"thumbnail"`
	Cover       string `gorm:"column:cover" json:"cover"`
	Sorts       int    `gorm:"column:sorts" json:"sorts"`
}

func (News) TableName() string {
	return "news"
}

type NewsCategory struct {
	ID         uint   `gorm:"primaryKey;column:id" json:"id"`
	Name       string `gorm:"column:name" json:"name"`
	Sorts      int    `gorm:"column:sorts" json:"sorts"`
	IsShow     int8   `gorm:"column:is_show" json:"is_show"`
	CreateTime int64  `gorm:"column:create_time" json:"create_time"`
	UpdateTime int64  `gorm:"column:update_time" json:"update_time"`
	SiteID     int8   `gorm:"column:site_id" json:"site_id"`
}

func (NewsCategory) TableName() string {
	return "news_category"
}

type NewsCopy struct {
	ID          uint   `gorm:"primaryKey;column:id" json:"id"`
	CID         uint   `gorm:"column:c_id" json:"c_id"`
	Lang        string `gorm:"column:lang" json:"lang"`
	Title       string `gorm:"column:title" json:"title"`
	Recommend   int8   `gorm:"column:recommend" json:"recommend"`
	Audit       int8   `gorm:"column:audit" json:"audit"`
	Display     int8   `gorm:"column:display" json:"display"`
	Discuss     int8   `gorm:"column:discuss" json:"discuss"`
	Author      string `gorm:"column:author" json:"author"`
	BrowseGrant int8   `gorm:"column:browse_grant" json:"browse_grant"`
	Keyword     string `gorm:"column:keyword" json:"keyword"`
	Abstract    string `gorm:"column:abstract" json:"abstract"`
	Content     string `gorm:"column:content" json:"content"`
	Views       int    `gorm:"column:views" json:"views"`
	CreateTime  int64  `gorm:"column:create_time" json:"create_time"`
	UpdateTime  int64  `gorm:"column:update_time" json:"update_time"`
	Thumbnail   string `gorm:"column:thumbnail" json:"thumbnail"`
	Cover       string `gorm:"column:cover" json:"cover"`
	Sorts       int    `gorm:"column:sorts" json:"sorts"`
}

func (NewsCopy) TableName() string {
	return "news_copy"
}

type PrizePool struct {
	ID             uint    `gorm:"primaryKey;column:id" json:"id"`
	Scene          int     `gorm:"column:scene" json:"scene"`
	RewardType     int8    `gorm:"column:reward_type" json:"reward_type"`
	RewardCurrency int     `gorm:"column:reward_currency" json:"reward_currency"`
	CurrencyType   int     `gorm:"column:currency_type" json:"currency_type"`
	RewardQty      float64 `gorm:"column:reward_qty" json:"reward_qty"`
	FromUserID     uint    `gorm:"column:from_user_id" json:"from_user_id"`
	ToUserID       uint    `gorm:"column:to_user_id" json:"to_user_id"`
	Status         int8    `gorm:"column:status" json:"status"`
	Sign           int     `gorm:"column:sign" json:"sign"`
	ExtraData      string  `gorm:"column:extra_data" json:"extra_data"`
	Memo           string  `gorm:"column:memo" json:"memo"`
	CreateTime     int64   `gorm:"column:create_time" json:"create_time"`
	ExpireTime     int64   `gorm:"column:expire_time" json:"expire_time"`
	ReceiveTime    int64   `gorm:"column:receive_time" json:"receive_time"`
	ErrorInfo      string  `gorm:"column:error_info" json:"error_info"`
}

func (PrizePool) TableName() string {
	return "prize_pool"
}

type Robot struct {
	ID              uint    `gorm:"primaryKey;column:id" json:"id"`
	CurrencyID      uint    `gorm:"column:currency_id" json:"currency_id"`
	LegalID         uint    `gorm:"column:legal_id" json:"legal_id"`
	BuyUserID       uint    `gorm:"column:buy_user_id" json:"buy_user_id"`
	SellUserID      uint    `gorm:"column:sell_user_id" json:"sell_user_id"`
	CreateTime      int64   `gorm:"column:create_time" json:"create_time"`
	Status          int     `gorm:"column:status" json:"status"`
	Second          int     `gorm:"column:second" json:"second"`
	Sell            int     `gorm:"column:sell" json:"sell"`
	Buy             int     `gorm:"column:buy" json:"buy"`
	NumberMax       float64 `gorm:"column:number_max" json:"number_max"`
	NumberMin       float64 `gorm:"column:number_min" json:"number_min"`
	FloatNumberDown float64 `gorm:"column:float_number_down" json:"float_number_down"`
	FloatNumberUp   float64 `gorm:"column:float_number_up" json:"float_number_up"`
	HuobiCurrency   string  `gorm:"column:huobi_currency" json:"huobi_currency"`
	Mult            string  `gorm:"column:mult" json:"mult"`
}

func (Robot) TableName() string {
	return "robot"
}

type RobotPlan struct {
	ID        uint    `gorm:"primaryKey;column:id" json:"id"`
	Itime     int64   `gorm:"column:itime" json:"itime"`
	Remark    string  `gorm:"column:remark" json:"remark"`
	Base      string  `gorm:"column:base" json:"base"`
	Target    string  `gorm:"column:target" json:"target"`
	FloatDown float64 `gorm:"column:float_down" json:"float_down"`
	FloatUp   float64 `gorm:"column:float_up" json:"float_up"`
	MaxPrice  float64 `gorm:"column:max_price" json:"max_price"`
	MinPrice  float64 `gorm:"column:min_price" json:"min_price"`
	Rid       int     `gorm:"column:rid" json:"rid"`
	Etime     int64   `gorm:"column:etime" json:"etime"`
}

func (RobotPlan) TableName() string {
	return "robot_plan"
}

type Seller struct {
	ID                uint    `gorm:"primaryKey;column:id" json:"id"`
	UserID            uint    `gorm:"column:user_id" json:"user_id"`
	Name              string  `gorm:"column:name" json:"name"`
	SellerBalance     float64 `gorm:"column:seller_balance" json:"seller_balance"`
	LockSellerBalance float64 `gorm:"column:lock_seller_balance" json:"lock_seller_balance"`
	WechatNickname    string  `gorm:"column:wechat_nickname" json:"wechat_nickname"`
	WechatAccount     string  `gorm:"column:wechat_account" json:"wechat_account"`
	AliNickname       string  `gorm:"column:ali_nickname" json:"ali_nickname"`
	AliAccount        string  `gorm:"column:ali_account" json:"ali_account"`
	BankID            uint    `gorm:"column:bank_id" json:"bank_id"`
	BankAccount       string  `gorm:"column:bank_account" json:"bank_account"`
	BankAddress       string  `gorm:"column:bank_address" json:"bank_address"`
	CreateTime        int64   `gorm:"column:create_time" json:"create_time"`
	CurrencyID        uint    `gorm:"column:currency_id" json:"currency_id"`
	Mobile            string  `gorm:"column:mobile" json:"mobile"`
	AlipayQrCode      string  `gorm:"column:alipay_qr_code" json:"alipay_qr_code"`
	WechatQrCode      string  `gorm:"column:wechat_qr_code" json:"wechat_qr_code"`
	Status            int     `gorm:"column:status" json:"status"`
}

func (Seller) TableName() string {
	return "seller"
}

type Setting struct {
	ID    uint   `gorm:"primaryKey;column:id" json:"id"`
	Name  string `gorm:"-" json:"name"` // 不映射到数据库列，避免冲突
	Key   string `gorm:"column:key;type:varchar(50)" json:"key"`
	Value string `gorm:"column:value;type:varchar(1000)" json:"value"`
	Notes string `gorm:"column:notes;type:varchar(255)" json:"notes"`
}

func (Setting) TableName() string {
	return "settings"
}

type Token struct {
	ID      uint   `gorm:"primaryKey;column:id" json:"id"`
	Token   string `gorm:"column:token" json:"token"`
	TimeOut int    `gorm:"column:time_out" json:"time_out"`
	UserID  uint   `gorm:"column:user_id" json:"user_id"`
}

func (Token) TableName() string {
	return "tokens"
}

type Transaction struct {
	ID         uint    `gorm:"primaryKey;column:id" json:"id"`
	FromUserID uint    `gorm:"column:from_user_id" json:"from_user_id"`
	Currency   int     `gorm:"column:currency" json:"currency"`
	LegalID    uint    `gorm:"column:legal_id" json:"legal_id"`
	ToUserID   uint    `gorm:"column:to_user_id" json:"to_user_id"`
	Type       int8    `gorm:"column:type" json:"type"`
	Number     float64 `gorm:"column:number" json:"number"`
	DealNumber float64 `gorm:"column:deal_number" json:"deal_number"`
	OrderValue float64 `gorm:"column:order_value" json:"order_value"` // 订单价值(USDT)
	Remarks    string  `gorm:"column:remarks" json:"remarks"`
	Time       int64   `gorm:"column:time" json:"time"`
	Status     int     `gorm:"column:status" json:"status"`
	Price      float64 `gorm:"column:price" json:"price"`
}

func (Transaction) TableName() string {
	return "transaction"
}

type TransactionComplete struct {
	ID         uint    `gorm:"primaryKey;column:id" json:"id"`
	Way        int8    `gorm:"column:way" json:"way"`
	UserID     uint    `gorm:"column:user_id" json:"user_id"`
	Price      float64 `gorm:"column:price" json:"price"`
	Number     float64 `gorm:"column:number" json:"number"`
	CreateTime int64   `gorm:"column:create_time" json:"create_time"`
	Currency   int     `gorm:"column:currency" json:"currency"`
	FromUserID uint    `gorm:"column:from_user_id" json:"from_user_id"`
	Legal      int     `gorm:"column:legal" json:"legal"`
}

func (TransactionComplete) TableName() string {
	return "transaction_complete"
}

type TransactionIn struct {
	ID         uint    `gorm:"primaryKey;column:id" json:"id"`
	UserID     uint    `gorm:"column:user_id" json:"user_id"`
	Price      float64 `gorm:"column:price" json:"price"`
	Number     float64 `gorm:"column:number" json:"number"`
	CreateTime int64   `gorm:"column:create_time" json:"create_time"`
	Currency   int     `gorm:"column:currency" json:"currency"`
	Legal      int     `gorm:"column:legal" json:"legal"`
}

func (TransactionIn) TableName() string {
	return "transaction_in"
}

type TransactionOut struct {
	ID         uint    `gorm:"primaryKey;column:id" json:"id"`
	UserID     uint    `gorm:"column:user_id" json:"user_id"`
	Price      float64 `gorm:"column:price" json:"price"`
	Number     float64 `gorm:"column:number" json:"number"`
	CreateTime int64   `gorm:"column:create_time" json:"create_time"`
	Currency   int     `gorm:"column:currency" json:"currency"`
	Legal      int     `gorm:"column:legal" json:"legal"`
}

func (TransactionOut) TableName() string {
	return "transaction_out"
}

type UserAlgebra struct {
	ID          uint      `gorm:"primaryKey;column:id" json:"id"`
	UserID      uint      `gorm:"column:user_id" json:"user_id"`
	TouchUserID uint      `gorm:"column:touch_user_id" json:"touch_user_id"`
	Algebra     int       `gorm:"column:algebra" json:"algebra"`
	Value       float64   `gorm:"column:value" json:"value"`
	Info        string    `gorm:"column:info" json:"info"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (UserAlgebra) TableName() string {
	return "user_algebra"
}

type UserCashInfo struct {
	ID             uint   `gorm:"primaryKey;column:id" json:"id"`
	UserID         uint   `gorm:"column:user_id" json:"user_id"`
	BankID         uint   `gorm:"column:bank_id" json:"bank_id"`
	BankName       string `gorm:"column:bank_name" json:"bank_name"`
	BankBranch     string `gorm:"column:bank_branch" json:"bank_branch"`
	BankAccount    string `gorm:"column:bank_account" json:"bank_account"`
	RealName       string `gorm:"column:real_name" json:"real_name"`
	AlipayAccount  string `gorm:"column:alipay_account" json:"alipay_account"`
	WechatNickname string `gorm:"column:wechat_nickname" json:"wechat_nickname"`
	WechatAccount  string `gorm:"column:wechat_account" json:"wechat_account"`
	CreateTime     int64  `gorm:"column:create_time" json:"create_time"`
	AlipayQrCode   string `gorm:"column:alipay_qr_code" json:"alipay_qr_code"`
	WechatQrCode   string `gorm:"column:wechat_qr_code" json:"wechat_qr_code"`
}

func (UserCashInfo) TableName() string {
	return "user_cash_info"
}

type UserChat struct {
	ID         uint   `gorm:"primaryKey;column:id" json:"id"`
	FromUserID uint   `gorm:"column:from_user_id" json:"from_user_id"`
	ToUserID   uint   `gorm:"column:to_user_id" json:"to_user_id"`
	Content    string `gorm:"column:content" json:"content"`
	Offline    int8   `gorm:"column:offline" json:"offline"`
	Type       string `gorm:"column:type" json:"type"`
	AddTime    int64  `gorm:"column:add_time" json:"add_time"`
}

func (UserChat) TableName() string {
	return "user_chat"
}

type UserLoginInfo struct {
	UserID         uint      `gorm:"column:user_id" json:"user_id"`
	LastIpaddr     string    `gorm:"column:last_ipaddr" json:"last_ipaddr"`
	LastIpaddrInfo string    `gorm:"column:last_ipaddr_info" json:"last_ipaddr_info"`
	LastLoginTime  time.Time `gorm:"column:last_login_time" json:"last_login_time"`
}

func (UserLoginInfo) TableName() string {
	return "user_login_info"
}

type UserProfiles struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	UserID     uint      `gorm:"column:user_id" json:"user_id"`
	Name       string    `gorm:"column:name" json:"name"`
	CardID     string    `gorm:"column:card_id" json:"card_id"`
	FrontPic   string    `gorm:"column:front_pic" json:"front_pic"`
	ReversePic string    `gorm:"column:reverse_pic" json:"reverse_pic"`
	HandPic    string    `gorm:"column:hand_pic" json:"hand_pic"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (UserProfiles) TableName() string {
	return "user_profiles"
}

type UserReal struct {
	ID           uint   `gorm:"primaryKey;column:id" json:"id"`
	UserID       uint   `gorm:"column:user_id" json:"user_id"`
	Name         string `gorm:"column:name" json:"name"`
	CardID       string `gorm:"column:card_id" json:"card_id"`
	Phone        string `gorm:"column:phone" json:"phone"`
	BankCard     string `gorm:"column:bank_card" json:"bank_card"`
	ReviewStatus int8   `gorm:"column:review_status" json:"review_status"` // 0:初始 1:待审核 2:已通过 3:已拒绝
	RejectReason string `gorm:"column:reject_reason" json:"reject_reason"` // 拒绝原因
	FrontPic     string `gorm:"column:front_pic" json:"front_pic"`
	ReversePic   string `gorm:"column:reverse_pic" json:"reverse_pic"`
	HandPic      string `gorm:"column:hand_pic" json:"hand_pic"`
	BankPic      string `gorm:"column:bank_pic" json:"bank_pic"`
	CreateTime   int64  `gorm:"column:create_time" json:"create_time"`
	ReviewTime   int64  `gorm:"column:review_time" json:"review_time"`
}

func (UserReal) TableName() string {
	return "user_real"
}

type UserRealCopy struct {
	ID           uint   `gorm:"primaryKey;column:id" json:"id"`
	UserID       uint   `gorm:"column:user_id" json:"user_id"`
	Name         string `gorm:"column:name" json:"name"`
	CardID       string `gorm:"column:card_id" json:"card_id"`
	Phone        string `gorm:"column:phone" json:"phone"`
	BankCard     string `gorm:"column:bank_card" json:"bank_card"`
	ReviewStatus int8   `gorm:"column:review_status" json:"review_status"`
	FrontPic     string `gorm:"column:front_pic" json:"front_pic"`
	ReversePic   string `gorm:"column:reverse_pic" json:"reverse_pic"`
	HandPic      string `gorm:"column:hand_pic" json:"hand_pic"`
	BankPic      string `gorm:"column:bank_pic" json:"bank_pic"`
	CreateTime   int64  `gorm:"column:create_time" json:"create_time"`
	ReviewTime   int64  `gorm:"column:review_time" json:"review_time"`
}

func (UserRealCopy) TableName() string {
	return "user_real_copy"
}

type User struct {
	ID                           uint    `gorm:"primaryKey;column:id" json:"id"`
	AreaCodeID                   uint    `gorm:"column:area_code_id" json:"area_code_id"`
	AccountNumber                string  `gorm:"column:account_number;type:varchar(255);index" json:"account_number"`
	Type                         int8    `gorm:"column:type" json:"type"`
	Phone                        string  `gorm:"column:phone;type:varchar(50);index" json:"phone"`
	AgentID                      uint    `gorm:"column:agent_id" json:"agent_id"`
	AgentNoteID                  uint    `gorm:"column:agent_note_id" json:"agent_note_id"`
	ParentID                     uint    `gorm:"column:parent_id" json:"parent_id"`
	Email                        string  `gorm:"column:email;type:varchar(255);index" json:"email"`
	Password                     string  `gorm:"column:password" json:"password"`
	PayPassword                  string  `gorm:"column:pay_password" json:"pay_password"`
	Time                         int64   `gorm:"column:time" json:"time"`
	HeadPortrait                 string  `gorm:"column:head_portrait" json:"head_portrait"`
	ExtensionCode                string  `gorm:"column:extension_code" json:"extension_code"`
	Status                       int     `gorm:"column:status" json:"status"`
	GesturePassword              string  `gorm:"column:gesture_password" json:"gesture_password"`
	IsAuth                       string  `gorm:"column:is_auth" json:"is_auth"`
	Nickname                     string  `gorm:"column:nickname" json:"nickname"`
	WalletAddress                string  `gorm:"column:wallet_address" json:"wallet_address"`
	IsBlacklist                  int8    `gorm:"column:is_blacklist" json:"is_blacklist"`
	ParentsPath                  string  `gorm:"column:parents_path" json:"parents_path"`
	PushStatus                   int     `gorm:"column:push_status" json:"push_status"`
	CandyNumber                  float64 `gorm:"column:candy_number" json:"candy_number"`
	ZhituiRealNumber             int     `gorm:"column:zhitui_real_number" json:"zhitui_real_number"`
	RealTeamnumber               uint    `gorm:"column:real_teamnumber" json:"real_teamnumber"`
	TopUpnumber                  float64 `gorm:"column:top_upnumber" json:"top_upnumber"`
	IsRealname                   int     `gorm:"column:is_realname" json:"is_realname"`
	IsAtelier                    int     `gorm:"column:is_atelier" json:"is_atelier"`
	NewIsrealTime                int64   `gorm:"column:new_isreal_time" json:"new_isreal_time"`
	TodayRealTeamnumber          int     `gorm:"column:today_real_teamnumber" json:"today_real_teamnumber"`
	TodayLegalDealCancelNum      int     `gorm:"column:today_LegalDealCancel_num" json:"today_LegalDealCancel_num"`
	LegalDealCancelNumUpdateTime int64   `gorm:"column:LegalDealCancel_num__update_time" json:"LegalDealCancel_num__update_time"`
	Risk                         int8    `gorm:"column:risk" json:"risk"`
	LockTime                     int64   `gorm:"column:lock_time" json:"lock_time"`
	Level                        int     `gorm:"column:level" json:"level"`
	Fund                         float64 `gorm:"column:fund" json:"fund"`
	IsService                    uint8   `gorm:"column:is_service" json:"is_service"`
	AgentPath                    string  `gorm:"column:agent_path" json:"agent_path"`
	StoreID                      uint    `gorm:"column:store_id" json:"store_id"`
	LastTime                     int64   `gorm:"column:last_time" json:"last_time"`
	LastLoginIP                  string  `gorm:"column:last_login_ip" json:"last_login_ip"`
	UserRemark                   string  `gorm:"column:user_remark" json:"user_remark"`
}

func (User) TableName() string {
	return "users"
}

type UsersInsurances struct {
	ID              uint      `gorm:"primaryKey;column:id" json:"id"`
	UserID          uint      `gorm:"column:user_id" json:"user_id"`
	InsuranceTypeID uint32    `gorm:"column:insurance_type_id" json:"insurance_type_id"`
	Amount          float64   `gorm:"column:amount" json:"amount"`
	InsuranceAmount float64   `gorm:"column:insurance_amount" json:"insurance_amount"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`
	YieldedAt       time.Time `gorm:"column:yielded_at" json:"yielded_at"`
	RescindedAt     time.Time `gorm:"column:rescinded_at" json:"rescinded_at"`
	RescindedType   uint8     `gorm:"column:rescinded_type" json:"rescinded_type"`
	Status          uint8     `gorm:"column:status" json:"status"`
	ClaimStatus     uint8     `gorm:"column:claim_status" json:"claim_status"`
}

func (UsersInsurances) TableName() string {
	return "users_insurances"
}

// ============================================================
// UserAssets - 用户资产表(JSON聚合存储)
// ============================================================
// 注意：UsersWallet 表已废弃并删除，所有资产数据统一使用 UserAssets 表

// CurrencyBalance 币种余额映射类型
type CurrencyBalance map[string]float64

// Scan 实现 sql.Scanner 接口
func (cb *CurrencyBalance) Scan(value interface{}) error {
	if value == nil {
		*cb = make(CurrencyBalance)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("type assertion to []byte failed")
	}

	if len(bytes) == 0 {
		*cb = make(CurrencyBalance)
		return nil
	}

	return json.Unmarshal(bytes, cb)
}

// Value 实现 driver.Valuer 接口
func (cb CurrencyBalance) Value() (driver.Value, error) {
	if cb == nil {
		return json.Marshal(make(map[string]float64))
	}
	return json.Marshal(cb)
}

// UserAssets 用户资产表
// 支持现货钱包(spot)和合约钱包(contract)分离
type UserAssets struct {
	ID         uint       `gorm:"primaryKey;column:id" json:"id"`
	UserID     uint       `gorm:"column:user_id;not null" json:"user_id"`
	WalletType WalletType `gorm:"column:wallet_type;type:varchar(20);not null;default:'spot'" json:"wallet_type"` // 索引在数据库迁移中创建

	// USDT主账户余额
	UsdtBalance float64 `gorm:"column:usdt_balance;type:decimal(20,8);not null;default:0" json:"usdt_balance"`
	UsdtLocked  float64 `gorm:"column:usdt_locked;type:decimal(20,8);not null;default:0" json:"usdt_locked"`

	// 各币种余额(JSON格式) - 使用TEXT存储以兼容MySQL 5.7以下版本
	CurrencyBalances CurrencyBalance `gorm:"column:currency_balances;type:text" json:"currency_balances"`
	CurrencyLocked   CurrencyBalance `gorm:"column:currency_locked;type:text" json:"currency_locked"`

	// 交割合约专用字段
	DeliveryMargin   float64 `gorm:"column:delivery_margin;type:decimal(20,8);default:0" json:"delivery_margin"`   // 交割合约保证金
	DeliveryPnL      float64 `gorm:"column:delivery_pnl;type:decimal(20,8);default:0" json:"delivery_pnl"`       // 交割合约盈亏

	// 统计字段
	TotalValueUsdt float64 `gorm:"column:total_value_usdt;type:decimal(20,8);not null;default:0;index" json:"total_value_usdt"`
	LastTradeTime  int64   `gorm:"column:last_trade_time;not null;default:0;index" json:"last_trade_time"`
	LastUpdateTime int64   `gorm:"column:last_update_time;not null;default:0" json:"last_update_time"`

	// 版本控制(乐观锁)
	Version int `gorm:"column:version;not null;default:0" json:"version"`

	CreateTime int64 `gorm:"column:create_time;not null" json:"create_time"`
	UpdateTime int64 `gorm:"column:update_time;not null" json:"update_time"`
}

func (UserAssets) TableName() string {
	return "user_assets"
}

// GetCurrencyBalance 获取指定币种余额
func (ua *UserAssets) GetCurrencyBalance(currencyName string) float64 {
	if ua.CurrencyBalances == nil {
		return 0
	}
	if balance, exists := ua.CurrencyBalances[currencyName]; exists {
		return balance
	}
	return 0
}

// SetCurrencyBalance 设置指定币种余额
func (ua *UserAssets) SetCurrencyBalance(currencyName string, balance float64) {
	if ua.CurrencyBalances == nil {
		ua.CurrencyBalances = make(CurrencyBalance)
	}
	ua.CurrencyBalances[currencyName] = balance
}

// AddCurrencyBalance 增加指定币种余额
func (ua *UserAssets) AddCurrencyBalance(currencyName string, amount float64) {
	if ua.CurrencyBalances == nil {
		ua.CurrencyBalances = make(CurrencyBalance)
	}
	currentBalance := ua.GetCurrencyBalance(currencyName)
	ua.CurrencyBalances[currencyName] = currentBalance + amount
}

// GetCurrencyLocked 获取指定币种锁定余额
func (ua *UserAssets) GetCurrencyLocked(currencyName string) float64 {
	if ua.CurrencyLocked == nil {
		return 0
	}
	if locked, exists := ua.CurrencyLocked[currencyName]; exists {
		return locked
	}
	return 0
}

// SetCurrencyLocked 设置指定币种锁定余额
func (ua *UserAssets) SetCurrencyLocked(currencyName string, locked float64) {
	if ua.CurrencyLocked == nil {
		ua.CurrencyLocked = make(CurrencyBalance)
	}
	ua.CurrencyLocked[currencyName] = locked
}

// LockCurrency 锁定指定币种余额
func (ua *UserAssets) LockCurrency(currencyName string, amount float64) error {
	availableBalance := ua.GetCurrencyBalance(currencyName)
	if availableBalance < amount {
		return errors.New("insufficient balance")
	}

	ua.SetCurrencyBalance(currencyName, availableBalance-amount)
	locked := ua.GetCurrencyLocked(currencyName)
	ua.SetCurrencyLocked(currencyName, locked+amount)
	return nil
}

// UnlockCurrency 解锁指定币种余额
func (ua *UserAssets) UnlockCurrency(currencyName string, amount float64) error {
	locked := ua.GetCurrencyLocked(currencyName)
	if locked < amount {
		return errors.New("insufficient locked balance")
	}

	ua.SetCurrencyLocked(currencyName, locked-amount)
	availableBalance := ua.GetCurrencyBalance(currencyName)
	ua.SetCurrencyBalance(currencyName, availableBalance+amount)
	return nil
}

type UsersWalletOut struct {
	ID               uint    `gorm:"primaryKey;column:id" json:"id"`
	UserID           uint    `gorm:"column:user_id" json:"user_id"`
	CurrencyID       uint    `gorm:"column:currency" json:"currency_id"` // 统一使用CurrencyID，映射到数据库currency列
	Address          string  `gorm:"column:address" json:"address"`      // 提现地址或银行卡号
	Number           float64 `gorm:"column:number" json:"number"`
	CreateTime       int64   `gorm:"column:create_time" json:"create_time"`
	Rate             float64 `gorm:"column:rate" json:"rate"`
	Status           int8    `gorm:"column:status" json:"status"`
	Notes            string  `gorm:"column:notes" json:"notes"`
	RealNumber       float64 `gorm:"column:real_number" json:"real_number"`
	Txid             string  `gorm:"column:txid" json:"txid"`
	Verificationcode string  `gorm:"column:verificationcode" json:"verificationcode"`
	UpdateTime       int64   `gorm:"column:update_time" json:"update_time"`
	ToAdddress       string  `gorm:"column:to_adddress" json:"to_adddress"`
	UsdtType         string  `gorm:"column:usdt_type" json:"usdt_type"`

	// 提现类型相关字段
	WithdrawType int8   `gorm:"column:withdraw_type;default:1" json:"withdraw_type"` // 提现类型:1=银行转账,2=区块链提币
	NetworkType  string `gorm:"column:network_type" json:"network_type"`             // 网络类型:TRC20/ERC20/BRC20/BEP20/Polygon
	ChainAddress string `gorm:"column:chain_address" json:"chain_address"`           // 区块链钱包地址

	// 提现快照信息（用于审核展示）
	RealName    string `gorm:"column:real_name" json:"real_name"`
	IDCard      string `gorm:"column:id_card" json:"id_card"`
	BankName    string `gorm:"column:bank_name" json:"bank_name"`
	BankBranch  string `gorm:"column:bank_branch" json:"bank_branch"`
	BankAccount string `gorm:"column:bank_account" json:"bank_account"`
}

func (UsersWalletOut) TableName() string {
	return "users_wallet_out"
}

type WalletAddressLog struct {
	ID         uint   `gorm:"primaryKey;column:id" json:"id"`
	Ctime      int64  `gorm:"column:ctime" json:"ctime"`
	OldAddress string `gorm:"column:old_address" json:"old_address"`
	NewAddress string `gorm:"column:new_address" json:"new_address"`
	UserID     uint   `gorm:"column:user_id" json:"user_id"`
	CurrencyID uint   `gorm:"column:currency_id" json:"currency_id"`
	ManagerID  uint   `gorm:"column:manager_id" json:"manager_id"`
}

func (WalletAddressLog) TableName() string {
	return "wallet_address_log"
}

type WalletLog struct {
	ID           uint    `gorm:"primaryKey;column:id" json:"id"`
	UserID       uint    `gorm:"column:user_id" json:"user_id"`
	FromUserID   uint    `gorm:"column:from_user_id" json:"from_user_id"`
	AccountLogID uint    `gorm:"column:account_log_id" json:"account_log_id"`
	WalletID     uint    `gorm:"column:wallet_id" json:"wallet_id"`
	BalanceType  int     `gorm:"column:balance_type" json:"balance_type"`
	LockType     uint8   `gorm:"column:lock_type" json:"lock_type"`
	Before       float64 `gorm:"column:before" json:"before"`
	Change       float64 `gorm:"column:change" json:"change"`
	After        float64 `gorm:"column:after" json:"after"`
	Memo         string  `gorm:"column:memo" json:"memo"`
	ExtraSign    int     `gorm:"column:extra_sign" json:"extra_sign"`
	ExtraData    string  `gorm:"column:extra_data" json:"extra_data"`
	CreateTime   int64   `gorm:"column:create_time" json:"create_time"`
	MemoEn       string  `gorm:"column:memo_en" json:"memo_en"`
	MemoSpa      string  `gorm:"column:memo_spa" json:"memo_spa"`
	MemoJp       string  `gorm:"column:memo_jp" json:"memo_jp"`
	MemoHk       string  `gorm:"column:memo_hk" json:"memo_hk"`
	MemoKr       string  `gorm:"column:memo_kr" json:"memo_kr"`
	Transfered   int     `gorm:"column:transfered" json:"transfered"`
}

func (WalletLog) TableName() string {
	return "wallet_log"
}

type Zhiya struct {
	ID           uint    `gorm:"primaryKey;column:id" json:"id"`
	CurrencyID   uint    `gorm:"column:currency_id" json:"currency_id"`
	CurrencyName string  `gorm:"column:currency_name" json:"currency_name"`
	Qixian       int     `gorm:"column:qixian" json:"qixian"`
	Lixi         float64 `gorm:"column:lixi" json:"lixi"`
	Min          float64 `gorm:"column:min" json:"min"`
	Px           int     `gorm:"column:px" json:"px"`
	Status       int     `gorm:"column:status" json:"status"`
	IsDel        int     `gorm:"column:is_del" json:"is_del"`
	CreateTime   int64   `gorm:"column:create_time" json:"create_time"`
	UpdateTime   int64   `gorm:"column:update_time" json:"update_time"`
	Mark         string  `gorm:"column:mark" json:"mark"`
}

func (Zhiya) TableName() string {
	return "zhiya"
}

type ZhiyaOrder struct {
	ID            uint    `gorm:"primaryKey;column:id" json:"id"`
	UserID        uint    `gorm:"column:user_id" json:"user_id"`
	AccountNumber string  `gorm:"column:account_number" json:"account_number"`
	BuyNum        float64 `gorm:"column:buy_num" json:"buy_num"`
	FanLixi       float64 `gorm:"column:fan_lixi" json:"fan_lixi"`
	LixiTotal     float64 `gorm:"column:lixi_total" json:"lixi_total"`
	Status        int     `gorm:"column:status" json:"status"`
	CreateTime    int64   `gorm:"column:create_time" json:"create_time"`
	EndTime       int64   `gorm:"column:end_time" json:"end_time"`
	UpdateTime    int64   `gorm:"column:update_time" json:"update_time"`
	LixiTime      int64   `gorm:"column:lixi_time" json:"lixi_time"`
	IsDel         int     `gorm:"column:is_del" json:"is_del"`
	Mark          string  `gorm:"column:mark" json:"mark"`
	Other         string  `gorm:"column:other" json:"other"`
	ZhiyaID       uint    `gorm:"column:zhiya_id" json:"zhiya_id"`
	CurrencyID    uint    `gorm:"column:currency_id" json:"currency_id"`
	CurrencyName  string  `gorm:"column:currency_name" json:"currency_name"`
	ZhiyaQixian   int     `gorm:"column:zhiya_qixian" json:"zhiya_qixian"`
	ZhiyiaLixi    float64 `gorm:"column:zhiyia_lixi" json:"zhiyia_lixi"`
	P1            int     `gorm:"column:p1" json:"p1"`
	P2            int     `gorm:"column:p2" json:"p2"`
	P1fan         float64 `gorm:"column:p1fan" json:"p1fan"`
	P2fan         float64 `gorm:"column:p2fan" json:"p2fan"`
	P3            int     `gorm:"column:p3" json:"p3"`
	P3fan         float64 `gorm:"column:p3fan" json:"p3fan"`
}

func (ZhiyaOrder) TableName() string {
	return "zhiya_order"
}

type Message struct {
	ID         uint   `gorm:"primaryKey;column:id" json:"id"`
	Title      string `gorm:"column:title" json:"title"`
	Content    string `gorm:"column:content" json:"content"`
	Type       int8   `gorm:"column:type" json:"type"`
	Priority   int8   `gorm:"column:priority" json:"priority"`
	SenderID   uint   `gorm:"column:sender_id" json:"sender_id"`
	SenderType int8   `gorm:"column:sender_type" json:"sender_type"`
	IsBatch    int8   `gorm:"column:is_batch" json:"is_batch"`
	Attachment string `gorm:"column:attachment" json:"attachment"`
	Status     int8   `gorm:"column:status" json:"status"`
	CreateTime int64  `gorm:"column:create_time" json:"create_time"`
	UpdateTime int64  `gorm:"column:update_time" json:"update_time"`
}

func (Message) TableName() string {
	return "messages"
}

type UserMessage struct {
	ID         uint  `gorm:"primaryKey;column:id" json:"id"`
	UserID     uint  `gorm:"column:user_id" json:"user_id"`
	MessageID  uint  `gorm:"column:message_id" json:"message_id"`
	IsRead     int8  `gorm:"column:is_read" json:"is_read"`
	IsDeleted  int8  `gorm:"column:is_deleted" json:"is_deleted"`
	ReadTime   int64 `gorm:"column:read_time" json:"read_time"`
	CreateTime int64 `gorm:"column:create_time" json:"create_time"`
}

func (UserMessage) TableName() string {
	return "user_messages"
}

type MessageTemplate struct {
	ID          uint   `gorm:"primaryKey;column:id" json:"id"`
	Code        string `gorm:"column:code" json:"code"`
	Title       string `gorm:"column:title" json:"title"`
	Content     string `gorm:"column:content" json:"content"`
	Type        int8   `gorm:"column:type" json:"type"`
	Description string `gorm:"column:description" json:"description"`
	Variables   string `gorm:"column:variables" json:"variables"`
	Status      int8   `gorm:"column:status" json:"status"`
	CreateTime  int64  `gorm:"column:create_time" json:"create_time"`
	UpdateTime  int64  `gorm:"column:update_time" json:"update_time"`
}

func (MessageTemplate) TableName() string {
	return "message_templates"
}

// ====================================
// 永续合约交易相关模型
// ====================================

// RiskKline 风控K线数据模型
// 用途: 存储风控插针K线数据，与真实K线耦合后推送给客户端显示
// 注意: 风控K线仅用于显示，不参与任何交易逻辑
type RiskKline struct {
	ID             uint    `gorm:"primaryKey;column:id" json:"id"`
	CurrencyID     uint    `gorm:"column:currency_id" json:"currency_id"`
	LegalID        uint    `gorm:"column:legal_id;default:3" json:"legal_id"`
	Timestamp      int64   `gorm:"column:timestamp" json:"timestamp"`
	Period         string  `gorm:"column:period;default:1min" json:"period"`
	Open           float64 `gorm:"column:open" json:"open"`
	High           float64 `gorm:"column:high" json:"high"`
	Low            float64 `gorm:"column:low" json:"low"`
	Close          float64 `gorm:"column:close" json:"close"`
	Volume         float64 `gorm:"column:volume" json:"volume"`
	IsSpike        int8    `gorm:"column:is_spike;default:1" json:"is_spike"`
	SpikeDirection int8    `gorm:"column:spike_direction;default:0" json:"spike_direction"` // 0=无方向, 1=向上, 2=向下
	TargetUserID   *uint   `gorm:"column:target_user_id" json:"target_user_id,omitempty"`   // NULL=全局
	Remarks        string  `gorm:"column:remarks" json:"remarks,omitempty"`
	CreatedAt      int64   `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      int64   `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (RiskKline) TableName() string {
	return "risk_kline"
}

// UserRiskControl 用户风控配置模型
// 用途: 存储用户级别的风控配置（亏损模式、强制爆仓等）
type UserRiskControl struct {
	ID                      uint    `gorm:"primaryKey;column:id" json:"id"`
	UserID                  uint    `gorm:"column:user_id;uniqueIndex" json:"user_id"`
	RiskMode                int8    `gorm:"column:risk_mode;default:0" json:"risk_mode"`                                 // 0=正常, 1=亏损模式, 2=强制爆仓
	ForceLossEnabled        int8    `gorm:"column:force_loss_enabled;default:0" json:"force_loss_enabled"`               // 是否启用强制亏损
	ForceLiquidationEnabled int8    `gorm:"column:force_liquidation_enabled;default:0" json:"force_liquidation_enabled"` // 是否启用强制爆仓
	LossRatio               float64 `gorm:"column:loss_ratio;default:100.00" json:"loss_ratio"`                          // 强制亏损比例(%)
	ApplyToContract         int8    `gorm:"column:apply_to_contract;default:1" json:"apply_to_contract"`                 // 是否应用于永续合约
	ApplyToSpot             int8    `gorm:"column:apply_to_spot;default:0" json:"apply_to_spot"`                         // 是否应用于现货
	Remarks                 string  `gorm:"column:remarks" json:"remarks,omitempty"`
	OperatorID              *uint   `gorm:"column:operator_id" json:"operator_id,omitempty"`
	CreatedAt               int64   `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt               int64   `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (UserRiskControl) TableName() string {
	return "user_risk_control"
}

// 风控模式常量
const (
	RiskModeNormal           int8 = 0 // 正常模式
	RiskModeForceLoss        int8 = 1 // 亏损模式
	RiskModeForceLiquidation int8 = 2 // 强制爆仓模式
)

// 订单类型常量
const (
	ContractOrderTypeMarket int8 = 1 // 市价单
	ContractOrderTypeLimit  int8 = 2 // 限价单
)

// 平仓类型常量
const (
	CloseTypeManual      int8 = 1 // 手动平仓
	CloseTypeLiquidation int8 = 2 // 爆仓
	CloseTypeTakeProfit  int8 = 3 // 止盈
	CloseTypeStopLoss    int8 = 4 // 止损
)

// 合约状态常量
const (
	ContractStatusPending int8 = 0 // 挂单中（限价单未成交）
	ContractStatusActive  int8 = 1 // 交易中（持仓中）
	ContractStatusClosing int8 = 2 // 平仓中
	ContractStatusClosed  int8 = 3 // 已平仓
	ContractStatusCancel  int8 = 4 // 已撤单
)

// ============================================================
// 充值模块
// ============================================================

// DepositAddress 充值地址配置表（全局统一）
type DepositAddress struct {
	ID         uint   `gorm:"primaryKey;column:id" json:"id"`
	Network    string `gorm:"column:network;size:20" json:"network"`  // TRC20/ERC20/BEP20/BRC20
	Address    string `gorm:"column:address;size:128" json:"address"` // 充值地址
	QrCode     string `gorm:"column:qr_code;size:255" json:"qr_code"` // 二维码图片路径
	Status     int8   `gorm:"column:status;default:1" json:"status"`  // 1启用 0禁用
	Sort       int    `gorm:"column:sort;default:0" json:"sort"`      // 排序
	CreateTime int64  `gorm:"column:create_time" json:"create_time"`
	UpdateTime int64  `gorm:"column:update_time" json:"update_time"`
}

func (DepositAddress) TableName() string {
	return "deposit_address"
}

// DepositOrder 充值订单表
type DepositOrder struct {
	ID          uint    `gorm:"primaryKey;column:id" json:"id"`
	OrderNo     string  `gorm:"column:order_no;size:32;uniqueIndex" json:"order_no"` // 订单号
	UserID      uint    `gorm:"column:user_id;index" json:"user_id"`
	Network     string  `gorm:"column:network;size:20" json:"network"`            // 充值网络
	Address     string  `gorm:"column:address;size:128" json:"address"`           // 充值地址
	Amount      float64 `gorm:"column:amount" json:"amount"`                      // 充值金额
	Screenshot  string  `gorm:"column:screenshot;size:255" json:"screenshot"`     // 转账截图路径
	Status      int8    `gorm:"column:status;default:0" json:"status"`            // 0待审核 1已通过 2已拒绝 3已取消 4已过期
	AdminID     uint    `gorm:"column:admin_id" json:"admin_id"`                  // 审核管理员ID
	AdminRemark string  `gorm:"column:admin_remark;size:255" json:"admin_remark"` // 审核备注
	CreateTime  int64   `gorm:"column:create_time" json:"create_time"`
	UpdateTime  int64   `gorm:"column:update_time" json:"update_time"`
	ReviewTime  int64   `gorm:"column:review_time" json:"review_time"` // 审核时间
	ExpireTime  int64   `gorm:"column:expire_time" json:"expire_time"` // 过期时间(24h)
}

func (DepositOrder) TableName() string {
	return "deposit_order"
}

// 充值订单状态常量
const (
	DepositStatusPending  int8 = 0 // 待审核
	DepositStatusApproved int8 = 1 // 已通过
	DepositStatusRejected int8 = 2 // 已拒绝
	DepositStatusCanceled int8 = 3 // 已取消
	DepositStatusExpired  int8 = 4 // 已过期
)

// 充值地址状态常量
const (
	DepositAddressEnabled  int8 = 1 // 启用
	DepositAddressDisabled int8 = 0 // 禁用
)

// ============================================================
// 钱包划转系统模型
// ============================================================

// WalletType 钱包类型枚举
type WalletType string

const (
	WalletTypeSpot      WalletType = "spot"      // 现货钱包
	WalletTypeContract WalletType = "contract"  // 永续合约钱包
	WalletTypeDelivery WalletType = "delivery"  // 交割合约钱包
	WalletTypeFund     WalletType = "fund"      // 资金钱包（仅存储，不参与交易）
)

// UserWallet 用户钱包资产表
type UserWallet struct {
	ID               uint64          `gorm:"primaryKey;column:id;autoIncrement" json:"id"`
	UserID           uint64          `gorm:"column:user_id;not null;index:idx_user_id" json:"user_id"`
	WalletType       WalletType      `gorm:"column:wallet_type;type:varchar(20);not null;index:idx_wallet_type" json:"wallet_type"`
	CurrencyID       uint64          `gorm:"column:currency_id;not null;default:1" json:"currency_id"`
	CurrencyName     string          `gorm:"column:currency_name;type:varchar(50);default:'USDT'" json:"currency_name"`
	AvailableBalance decimal.Decimal `gorm:"column:available_balance;type:decimal(20,8);not null;default:0" json:"available_balance"`
	LockedBalance    decimal.Decimal `gorm:"column:locked_balance;type:decimal(20,8);not null;default:0" json:"locked_balance"`
	Version          int             `gorm:"column:version;not null;default:0" json:"version"`
	LastTradeTime    int64           `gorm:"column:last_trade_time;not null;default:0" json:"last_trade_time"`
	CreateTime       int64           `gorm:"column:create_time;not null" json:"create_time"`
	UpdateTime       int64           `gorm:"column:update_time;not null;index:idx_update_time" json:"update_time"`
}

func (UserWallet) TableName() string {
	return "user_wallets"
}

// ============================================================
// FundWallet - 资金钱包专用模型
// 特点：
// 1. 不参与交易，仅用于资金存储
// 2. 支持与现货、合约、交割合约账户自由划转
// 3. 提供更清晰的资金管理分离
// ============================================================

// FundWallet 资金钱包表
type FundWallet struct {
	ID               uint64          `gorm:"primaryKey;column:id;autoIncrement" json:"id"`
	UserID           uint64          `gorm:"column:user_id;not null;index:idx_fund_user_id" json:"user_id"`
	CurrencyID       uint64          `gorm:"column:currency_id;not null;default:1" json:"currency_id"`
	CurrencyName     string          `gorm:"column:currency_name;type:varchar(50);default:'USDT'" json:"currency_name"`
	AvailableBalance decimal.Decimal `gorm:"column:available_balance;type:decimal(20,8);not null;default:0" json:"available_balance"`
	LockedBalance    decimal.Decimal `gorm:"column:locked_balance;type:decimal(20,8);not null;default:0" json:"locked_balance"`
	TotalDeposit     decimal.Decimal `gorm:"column:total_deposit;type:decimal(20,8);not null;default:0" json:"total_deposit"`     // 累计充值
	TotalWithdraw    decimal.Decimal `gorm:"column:total_withdraw;type:decimal(20,8);not null;default:0" json:"total_withdraw"`    // 累计提现
	Version          int             `gorm:"column:version;not null;default:0" json:"version"`
	LastTradeTime    int64           `gorm:"column:last_trade_time;not null;default:0" json:"last_trade_time"`
	CreateTime       int64           `gorm:"column:create_time;not null" json:"create_time"`
	UpdateTime       int64           `gorm:"column:update_time;not null;index:idx_fund_update_time" json:"update_time"`
}

func (FundWallet) TableName() string {
	return "fund_wallets"
}

// GetTotalBalance 获取总余额
func (fw *FundWallet) GetTotalBalance() decimal.Decimal {
	return fw.AvailableBalance.Add(fw.LockedBalance)
}

// FundWalletTransfer 资金钱包划转记录表
type FundWalletTransfer struct {
	ID              uint64          `gorm:"primaryKey;column:id;autoIncrement" json:"id"`
	TransferNo      string          `gorm:"column:transfer_no;type:varchar(64);not null;uniqueIndex:uk_fund_transfer_no" json:"transfer_no"`
	UserID          uint64          `gorm:"column:user_id;not null;index:idx_fund_transfer_user" json:"user_id"`
	TransferType    int8            `gorm:"column:transfer_type;not null;index:idx_fund_transfer_type" json:"transfer_type"` // 1=充值到资金钱包,2=从资金钱包提走,3=转入现货,4=从现货转入,5=转入合约,6=从合约转入,7=转入交割,8=从交割转入
	FromWalletType  WalletType      `gorm:"column:from_wallet_type;type:varchar(20)" json:"from_wallet_type"`   // 转出钱包类型
	ToWalletType    WalletType      `gorm:"column:to_wallet_type;type:varchar(20)" json:"to_wallet_type"`     // 转入钱包类型
	CurrencyID      uint64          `gorm:"column:currency_id;not null;default:1" json:"currency_id"`
	CurrencyName    string          `gorm:"column:currency_name;type:varchar(50);default:'USDT'" json:"currency_name"`
	Amount          decimal.Decimal `gorm:"column:amount;type:decimal(20,8);not null" json:"amount"`
	Fee             decimal.Decimal `gorm:"column:fee;type:decimal(20,8);not null;default:0" json:"fee"`
	BalanceBefore   decimal.Decimal `gorm:"column:balance_before;type:decimal(20,8);not null" json:"balance_before"` // 操作前余额
	BalanceAfter    decimal.Decimal `gorm:"column:balance_after;type:decimal(20,8);not null" json:"balance_after"`  // 操作后余额
	Status          int8            `gorm:"column:status;not null;default:0;index:idx_fund_transfer_status" json:"status"` // 0=待处理,1=成功,2=失败
	Remark          string          `gorm:"column:remark;type:varchar(500)" json:"remark"`
	ClientIP        string          `gorm:"column:client_ip;type:varchar(50)" json:"client_ip"`
	CreatedTime     int64           `gorm:"column:created_time;not null;index:idx_fund_transfer_created" json:"created_time"`
	CompletedTime   int64           `gorm:"column:completed_time;not null;default:0" json:"completed_time"`
}

func (FundWalletTransfer) TableName() string {
	return "fund_wallet_transfers"
}

// 划转类型常量
const (
	FundTransferDeposit     = 1 // 充值到资金钱包
	FundTransferWithdraw    = 2 // 从资金钱包提走
	FundTransferToSpot     = 3 // 转入现货
	FundTransferFromSpot   = 4 // 从现货转入
	FundTransferToContract = 5 // 转入合约
	FundTransferFromContract = 6 // 从合约转入
	FundTransferToDelivery = 7 // 转入交割
	FundTransferFromDelivery = 8 // 从交割转入
)

// 划转状态常量
const (
	FundTransferPending = 0 // 待处理
	FundTransferSuccess = 1 // 成功
	FundTransferFailed  = 2 // 失败
)

// DeliveryContract 交割合约配置表
type DeliveryContract struct {
	ID          uint    `gorm:"primaryKey;column:id" json:"id"`
	Symbol      string  `gorm:"column:symbol;uniqueIndex" json:"symbol"`           // 合约代码
	BaseAsset   string  `gorm:"column:base_asset" json:"base_asset"`             // 基础资产
	QuoteAsset  string  `gorm:"column:quote_asset" json:"quote_asset"`           // 计价资产
	ContractSize float64 `gorm:"column:contract_size" json:"contract_size"`       // 合约乘数
	MinPrice    float64 `gorm:"column:min_price" json:"min_price"`               // 最小价格变动单位
	MaxLeverage int     `gorm:"column:max_leverage" json:"max_leverage"`         // 最大杠杆倍数
	Status      int8    `gorm:"column:status" json:"status"`                    // 1=启用,0=禁用
	
	// 交割相关
	DeliveryDate time.Time `gorm:"column:delivery_date" json:"delivery_date"`     // 交割日期
	SettleType  int8     `gorm:"column:settle_type" json:"settle_type"`       // 交割方式: 1=现金, 2=实物
	
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time" json:"update_time"`
}

func (DeliveryContract) TableName() string {
	return "delivery_contracts"
}

// DeliveryPosition 交割合约持仓表
type DeliveryPosition struct {
	ID           uint    `gorm:"primaryKey;column:id" json:"id"`
	UserID       uint    `gorm:"column:user_id;index" json:"user_id"`
	ContractID   uint    `gorm:"column:contract_id;index" json:"contract_id"`
	Symbol       string  `gorm:"column:symbol" json:"symbol"`
	Side         int8    `gorm:"column:side" json:"side"`                       // 1=做多,2=做空
	Size         float64 `gorm:"column:size" json:"size"`                       // 持仓数量
	EntryPrice   float64 `gorm:"column:entry_price" json:"entry_price"`         // 开仓价格
	Leverage     int     `gorm:"column:leverage" json:"leverage"`              // 杠杆倍数
	Margin       float64 `gorm:"column:margin" json:"margin"`                   // 保证金
	
	// 当前价格和盈亏
	CurrentPrice float64 `gorm:"column:current_price" json:"current_price"`    // 当前标记价格
	UnrealizedPnL float64 `gorm:"column:unrealized_pnl" json:"unrealized_pnl"` // 未实现盈亏
	
	// 止盈止损
	TakeProfitPrice *float64 `gorm:"column:take_profit_price" json:"take_profit_price"`
	StopLossPrice   *float64 `gorm:"column:stop_loss_price" json:"stop_loss_price"`
	
	// 状态管理
	Status         int8     `gorm:"column:status" json:"status"`               // 1=持仓,0=已平仓
	DeliveryStatus int8     `gorm:"column:delivery_status" json:"delivery_status"` // 交割状态
	
	CreateTime time.Time `gorm:"column:created_at" json:"created_at"`
	UpdateTime time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (DeliveryPosition) TableName() string {
	return "delivery_positions"
}

// GetTotalBalance 获取总余额
func (uw *UserWallet) GetTotalBalance() decimal.Decimal {
	return uw.AvailableBalance.Add(uw.LockedBalance)
}

// WalletTransferRecord 钱包划转记录表
type WalletTransferRecord struct {
	ID             uint64          `gorm:"primaryKey;column:id;autoIncrement" json:"id"`
	TransferNo     string          `gorm:"column:transfer_no;type:varchar(64);not null;uniqueIndex:uk_transfer_no" json:"transfer_no"`
	UserID         uint64          `gorm:"column:user_id;not null;index:idx_user_id" json:"user_id"`
	FromWalletType WalletType      `gorm:"column:from_wallet_type;type:varchar(20);not null" json:"from_wallet_type"`
	ToWalletType   WalletType      `gorm:"column:to_wallet_type;type:varchar(20);not null" json:"to_wallet_type"`
	CurrencyID     uint64          `gorm:"column:currency_id;not null;default:1" json:"currency_id"`
	CurrencyName   string          `gorm:"column:currency_name;type:varchar(50);default:'USDT'" json:"currency_name"`
	Amount         decimal.Decimal `gorm:"column:amount;type:decimal(20,8);not null" json:"amount"`
	Fee            decimal.Decimal `gorm:"column:fee;type:decimal(20,8);not null;default:0" json:"fee"`
	Status         int8            `gorm:"column:status;not null;default:1;index:idx_status" json:"status"`
	Remark         string          `gorm:"column:remark;type:varchar(500)" json:"remark"`
	ClientIP       string          `gorm:"column:client_ip;type:varchar(45)" json:"client_ip"`
	CreatedTime    int64           `gorm:"column:created_time;not null;index:idx_created_time" json:"created_time"`
	CompletedTime  *int64          `gorm:"column:completed_time" json:"completed_time"`
	ErrorMsg       string          `gorm:"column:error_msg;type:varchar(500)" json:"error_msg"`
}

func (WalletTransferRecord) TableName() string {
	return "wallet_transfer_records"
}

// 划转记录状态常量
const (
	TransferStatusPending int8 = 1 // 处理中
	TransferStatusSuccess int8 = 2 // 成功
	TransferStatusFailed  int8 = 3 // 失败
)

// WalletAdjustmentRecord 钱包余额调整记录表
type WalletAdjustmentRecord struct {
	ID             uint64          `gorm:"primaryKey;column:id;autoIncrement" json:"id"`
	AdjustmentNo   string          `gorm:"column:adjustment_no;type:varchar(64);not null;uniqueIndex:uk_adjustment_no" json:"adjustment_no"`
	UserID         uint64          `gorm:"column:user_id;not null;index:idx_user_id" json:"user_id"`
	WalletType     WalletType      `gorm:"column:wallet_type;type:varchar(20);not null;index:idx_wallet_type" json:"wallet_type"`
	CurrencyID     uint64          `gorm:"column:currency_id;not null;default:1" json:"currency_id"`
	CurrencyName   string          `gorm:"column:currency_name;type:varchar(50);default:'USDT'" json:"currency_name"`
	AdjustmentType int8            `gorm:"column:adjustment_type;not null" json:"adjustment_type"`
	Amount         decimal.Decimal `gorm:"column:amount;type:decimal(20,8);not null" json:"amount"`
	BalanceBefore  decimal.Decimal `gorm:"column:balance_before;type:decimal(20,8);not null" json:"balance_before"`
	BalanceAfter   decimal.Decimal `gorm:"column:balance_after;type:decimal(20,8);not null" json:"balance_after"`
	OperatorID     uint64          `gorm:"column:operator_id;not null;index:idx_operator_id" json:"operator_id"`
	OperatorName   string          `gorm:"column:operator_name;type:varchar(100);not null" json:"operator_name"`
	Reason         string          `gorm:"column:reason;type:varchar(500);not null" json:"reason"`
	Remark         string          `gorm:"column:remark;type:varchar(500)" json:"remark"`
	Status         int8            `gorm:"column:status;not null;default:1" json:"status"`
	CreatedTime    int64           `gorm:"column:created_time;not null;index:idx_created_time" json:"created_time"`
	CompletedTime  *int64          `gorm:"column:completed_time" json:"completed_time"`
}

func (WalletAdjustmentRecord) TableName() string {
	return "wallet_adjustment_records"
}

// 调整记录状态常量
const (
	AdjustmentStatusPending int8 = 0 // 处理中
	AdjustmentStatusSuccess int8 = 1 // 成功
	AdjustmentStatusFailed  int8 = 2 // 失败
)

// 调整类型常量
const (
	AdjustmentTypeIncrease int8 = 1 // 增加
	AdjustmentTypeDecrease int8 = 2 // 减少
)
