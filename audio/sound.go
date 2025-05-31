package audio

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
)

const (
	MaxChannels   = 8               // 最大チャンネル数
	SoundDuration = 1 * time.Second // 効果音の再生時間
)

type SoundEffect struct {
	players    []*audio.Player // 複数のプレーヤーを保持
	volume     float64
	pan        float64       // -1.0 (左) から 1.0 (右)
	isPlaying  []bool        // 各チャンネルの再生状態
	stopTimers []*time.Timer // 各チャンネルの停止タイマー
	mutex      sync.Mutex
}

type SoundManager struct {
	context *audio.Context
	sounds  map[string]*SoundEffect
	mutex   sync.Mutex
}

var (
	instance *SoundManager
	once     sync.Once
)

// GetInstance はSoundManagerのシングルトンインスタンスを返します
func GetInstance() *SoundManager {
	once.Do(func() {
		instance = &SoundManager{
			context: audio.NewContext(44100),
			sounds:  make(map[string]*SoundEffect),
		}
	})
	return instance
}

// LoadSound は効果音を読み込みます
func (sm *SoundManager) LoadSound(name string, reader io.Reader) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// ファイルの内容をメモリに読み込む
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, reader); err != nil {
		return err
	}

	// 複数のプレーヤーを作成
	players := make([]*audio.Player, MaxChannels)
	isPlaying := make([]bool, MaxChannels)
	stopTimers := make([]*time.Timer, MaxChannels)

	for i := 0; i < MaxChannels; i++ {
		// MP3ファイルをデコード
		decoded, err := mp3.Decode(sm.context, bytes.NewReader(buf.Bytes()))
		if err != nil {
			return err
		}

		// プレーヤーを作成（ループなし）
		player, err := sm.context.NewPlayer(decoded)
		if err != nil {
			return err
		}

		players[i] = player
		isPlaying[i] = false
		stopTimers[i] = nil
	}

	// サウンドエフェクトを作成
	sound := &SoundEffect{
		players:    players,
		volume:     1.0,
		pan:        0.0,
		isPlaying:  isPlaying,
		stopTimers: stopTimers,
	}

	sm.sounds[name] = sound
	return nil
}

// Play は指定された効果音を再生します
func (sm *SoundManager) Play(name string) {
	sm.mutex.Lock()
	sound, exists := sm.sounds[name]
	sm.mutex.Unlock()

	if !exists {
		return
	}

	sound.mutex.Lock()
	defer sound.mutex.Unlock()

	// 使用可能なチャンネルを探す
	channel := -1
	for i := 0; i < MaxChannels; i++ {
		if !sound.isPlaying[i] {
			channel = i
			break
		}
	}

	// 使用可能なチャンネルがない場合は、最初のチャンネルを使用
	if channel == -1 {
		channel = 0
		// 現在再生中の音を停止
		if sound.stopTimers[channel] != nil {
			sound.stopTimers[channel].Stop()
		}
		sound.players[channel].Pause()
		sound.players[channel].Rewind()
	}

	// 新しい音を再生
	sound.players[channel].Play()
	sound.isPlaying[channel] = true

	// 既存のタイマーがあれば停止
	if sound.stopTimers[channel] != nil {
		sound.stopTimers[channel].Stop()
	}

	// 新しい停止タイマーを設定
	sound.stopTimers[channel] = time.AfterFunc(SoundDuration, func() {
		sound.mutex.Lock()
		defer sound.mutex.Unlock()

		if sound.isPlaying[channel] {
			sound.players[channel].Pause()
			sound.players[channel].Rewind()
			sound.isPlaying[channel] = false
			sound.stopTimers[channel] = nil
		}
	})
}

// SetVolume は効果音の音量を設定します
func (sm *SoundManager) SetVolume(name string, volume float64) {
	sm.mutex.Lock()
	sound, exists := sm.sounds[name]
	sm.mutex.Unlock()

	if !exists {
		return
	}

	sound.mutex.Lock()
	defer sound.mutex.Unlock()

	sound.volume = volume
	for _, player := range sound.players {
		player.SetVolume(volume)
	}
}

// SetPan は効果音の左右位置を設定します
func (sm *SoundManager) SetPan(name string, pan float64) {
	sm.mutex.Lock()
	sound, exists := sm.sounds[name]
	sm.mutex.Unlock()

	if !exists {
		return
	}

	sound.mutex.Lock()
	defer sound.mutex.Unlock()

	sound.pan = pan
	// TODO: パンニングの実装
}

// Stop は効果音の再生を停止します
func (sm *SoundManager) Stop(name string) {
	sm.mutex.Lock()
	sound, exists := sm.sounds[name]
	sm.mutex.Unlock()

	if !exists {
		return
	}

	sound.mutex.Lock()
	defer sound.mutex.Unlock()

	for i := 0; i < MaxChannels; i++ {
		if sound.isPlaying[i] {
			if sound.stopTimers[i] != nil {
				sound.stopTimers[i].Stop()
				sound.stopTimers[i] = nil
			}
			sound.players[i].Pause()
			sound.players[i].Rewind()
			sound.isPlaying[i] = false
		}
	}
}
