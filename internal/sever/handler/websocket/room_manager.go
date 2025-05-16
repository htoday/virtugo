package websocket

import "sync"

type RoomManager struct {
	rooms sync.Map // key: roomID, value: *Room
}

var Rooms = &RoomManager{}

func (rm *RoomManager) GetRoom(roomID string) (*Room, bool) {
	room, ok := rm.rooms.Load(roomID)
	if !ok {
		return nil, false
	}
	return room.(*Room), true
}

// 注册房间

func (m *RoomManager) Add(roomID string, r *Room) {
	m.rooms.Store(roomID, r)
}

// 删除房间：先关闭资源，再从管理器移除

func (m *RoomManager) Delete(roomID string) {
	v, ok := m.rooms.Load(roomID)
	if !ok {
		return
	}
	r := v.(*Room)
	r.CloseRoom() // 彻底清理资源
	m.rooms.Delete(roomID)
}
