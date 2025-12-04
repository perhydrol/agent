package domain

import (
	"sync"

	"github.com/bwmarrin/snowflake"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"go.uber.org/zap"
)

var (
	node *snowflake.Node
	once sync.Once
)

// InitIDGenerator 初始化雪花算法节点
func InitIDGenerator(nodeID int64) {
	once.Do(func() {
		var err error
		node, err = snowflake.NewNode(nodeID)
		if err != nil {
			logger.Log.Panic("failed to initialize snowflake node.", zap.Error(err))
		}
	})
}

// GenID 生成 int64 ID
func GenID() int64 {
	if node == nil {
		// 防止忘记初始化导致 panic，兜底给一个 node 1
		InitIDGenerator(1)
	}
	return node.Generate().Int64()
}
