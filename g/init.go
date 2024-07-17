package g

var Config = new(Configuration)

type Configuration struct {
	// 根据名称获取一个哈希值，根据哈希值匹配协程，默认协程的数量为10，控制并发数
	ConsistentNum int
}

func (c *Configuration) Name() string {
	return "g"
}

func (c *Configuration) Init() error {
	if c.ConsistentNum == 0 {
		c.ConsistentNum = 10
	}
	return nil
}

func (c *Configuration) Close() error {
	return nil
}
