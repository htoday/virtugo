package websocket

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"sync"
	"time"
	"virtugo/internal/config"
	"virtugo/internal/dao"
	"virtugo/internal/sever/llm"
	"virtugo/logs"
)

type Room struct {
	RoomID string `json:"room_id"`
	//Agents   map[string]*llm.ChatChain `json:"agents"`
	Agents        sync.Map
	ttsQueue      chan llm.TTSRequest
	MsgID         int64
	PlayDoneMsgID int64
	PlayDone      chan int64

	talkMu     sync.Mutex
	talkCancel context.CancelFunc
}

func NewRoom(ttsQueue chan llm.TTSRequest) *Room {
	return &Room{
		RoomID:        uuid.New().String(),
		Agents:        sync.Map{},
		ttsQueue:      ttsQueue,
		MsgID:         0,
		PlayDone:      make(chan int64, 1000),
		PlayDoneMsgID: -1,
	}
}
func (r *Room) AddAI(roleName string) {
	if _, loaded := r.Agents.Load(roleName); loaded {
		return
	}
	var agent llm.ChatChain
	agent.InitChain(r.ttsQueue, roleName)
	r.Agents.Store(roleName, &agent)
}
func (r *Room) ExitAI(roleName string) {
	if v, ok := r.Agents.Load(roleName); ok {
		v.(*llm.ChatChain).StopStream()
		r.Agents.Delete(roleName)
	}
}
func (r *Room) agentCount() int {
	cnt := 0
	r.Agents.Range(func(_, _ any) bool { cnt++; return true })
	return cnt
}

func (r *Room) Speak(input map[string]any) {

	//if r.talkCancel != nil {
	//	r.talkCancel()
	//}
	r.StopTalk()
	r.talkMu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	r.talkCancel = cancel
	r.talkMu.Unlock()

	go func() {
		r.PlayDoneMsgID = r.MsgID - 1
		if err := r.startTalk(ctx, input); err != nil {
			logs.Logger.Error("startTalk 出错", zap.Error(err))
		}
	}()
}

func (r *Room) startTalk(ctx context.Context, input map[string]any) error {
	count := r.agentCount()
	if count == 0 {
		return errors.New("没有可用的AI")
	}
	if count == 1 {
		input["chat_type"] = "single"
		// 收集所有成员
		members := ""
		r.Agents.Range(func(k, _ any) bool {
			members += k.(string) + ","
			return true
		})
		input["group_member"] = members
		var errOut error
		r.Agents.Range(func(k, v any) bool {
			input["ai_name"] = k
			if _, err := v.(*llm.ChatChain).Stream(input); err != nil {
				errOut = err
				return false
			}
			return true
		})
		return errOut
	}
	// 群聊
	// 在 StartTalk 里新增一个等待 ACK 的辅助函数
	waitAck := func(msgID int64) {
		for {
			preGenerateAmount := config.Cfg.PreGenerateAmount
			if r.PlayDoneMsgID+int64(preGenerateAmount)+1 >= msgID {
				return
			}
			select {
			case <-ctx.Done():
				return
			case ackID := <-r.PlayDone:
				r.PlayDoneMsgID = ackID //playDoneMsgID只会比ackID小
				if ackID+int64(preGenerateAmount)+1 >= msgID {
					//+2代表提前生成1条消息
					// 匹配成功，退出等待
					return
				}
				logs.Logger.Warn("播放完成信号不匹配，继续重试", zap.Int64("expected", msgID), zap.Int64("got", ackID))
			case <-time.After(time.Second * 60):
				logs.Logger.Warn("等待播放完成超时，停止重试", zap.Int64("msgID", msgID))
				return
			}
		}
	}
	sessionID, ok := input["session_id"].(int)
	if !ok {
		logs.Logger.Error("没有提供有效的会话ID")
		return errors.New("没有提供有效的会话ID")
	}
	username, _ := input["username"].(string)
	inputText, ok := input["question"].(string)
	if !ok {
		logs.Logger.Error("没有提供有效的输入文本")
		return errors.New("没有提供有效的输入文本")
	}
	_, _ = llm.SaveMessageToSession(dao.SqliteDB, sessionID, "user", username, inputText, 0)
	input["chat_type"] = "group"
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		// 收集所有成员
		members := ""
		r.Agents.Range(func(k, _ any) bool {
			members += k.(string) + ","
			return true
		})
		input["group_member"] = members

		var errOut error
		r.Agents.Range(func(k, v any) bool {
			logs.Logger.Info("现在轮到", zap.String("agent", k.(string)))
			input["ai_name"] = k
			input["msg_id"] = r.MsgID
			r.MsgID++
			if _, err := v.(*llm.ChatChain).Stream(input); err != nil {
				errOut = err
				return false
			}
			//战术停顿
			time.Sleep(time.Second * 1)
			// 等待前端的播放完成信号
			waitAck(r.MsgID)

			return true
		})
		if errOut != nil {
			return errOut
		}
	}
}

func (r *Room) StopTalk() {
	// 取消当前的对话
	r.talkMu.Lock()
	if r.talkCancel != nil {
		r.talkCancel()
	}
	r.talkCancel = nil
	r.talkMu.Unlock()
	//取消正在进行的生成
	r.Agents.Range(func(_, v any) bool {
		v.(*llm.ChatChain).StopStream()
		return true
	})
	// 清空 ttsQueue 中的所有待处理请求
drainLoop:
	for {
		select {
		case <-r.ttsQueue:
			// 丢弃
		default:
			break drainLoop
		}
	}
	//重置播放完成消息ID 防止下次信号不匹配
	r.PlayDoneMsgID = r.MsgID - 1

}

func (r *Room) CloseRoom() {
	r.StopTalk()
	if r.PlayDone != nil {
		close(r.PlayDone)
	}
	logs.Logger.Info("关闭房间", zap.String("room_id", r.RoomID))

}
