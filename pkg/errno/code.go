package errno

// 1. 通用错误 (10000 - 19999)
var (
	OK                = New(0, "success")
	ErrInternalServer = New(10001, "服务器内部错误")
	ErrBadRequest     = New(10002, "参数错误")
	ErrUnauthorized   = New(10003, "未授权")
	ErrNotFound       = New(10004, "资源不存在")
)

// 2. 订单模块错误 (20000 - 29999)
var (
	ErrOrderNotFound    = New(20001, "订单不存在")
	ErrOrderCannotPay   = New(20002, "订单状态不可支付")
	ErrOrderPriceChange = New(20003, "商品价格已变动，请重新下单")
	ErrOptimisticLock   = New(20004, "数据已更新，请刷新重试")
)

// 3. 仓储错误 (30000 - 39999)
var (
	ErrRepoDB         = New(30001, "数据库访问错误")
	ErrRepoTypeAssert = New(30002, "类型断言失败")
)

// 4. 缓存错误 (40000 - 49999)
var (
	ErrCacheGetFailed           = New(40001, "缓存读取失败")
	ErrCacheSetFailed           = New(40002, "缓存写入失败")
	ErrCacheDelFailed           = New(40003, "缓存删除失败")
	ErrCacheMarshalFailed       = New(40004, "缓存数据序列化失败")
	ErrCacheUnmarshalFailed     = New(40005, "缓存数据反序列化失败")
	ErrCacheLockerObtainFailed  = New(40006, "缓存锁获取失败")
	ErrCacheLockerReleaseFailed = New(40007, "缓存锁释放失败")
	ErrCacheLockerRefreshFailed = New(40008, "缓存锁续期失败")
	ErrCacheListParseFailed     = New(40009, "缓存列表解析失败")
)
