package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"virtugo/internal/config"
)

func GetSetting(c *gin.Context) {
	// 复制一份配置
	var safeCopy config.Config // 结构体类型

	// 深拷贝
	data, _ := json.Marshal(config.Cfg)
	json.Unmarshal(data, &safeCopy)

	// 清空敏感字段
	for key, mc := range safeCopy.Models {
		mc.ModelInfo.APIKey = ""
		mc.TTS.FishAudioKey = ""
		safeCopy.Models[key] = mc
	}

	// 返回副本
	c.JSON(http.StatusOK, safeCopy)
}
