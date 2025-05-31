package audio

import (
	"os"
)

// Initialize は効果音システムを初期化します
func Initialize() error {
	soundManager := GetInstance()

	// 効果音ファイルを読み込む
	shootSound, err := os.Open("assets/audio/se/SNES-Shooter02-01(Shoot).mp3")
	if err != nil {
		return err
	}

	// 効果音を登録
	if err := soundManager.LoadSound("shoot", shootSound); err != nil {
		shootSound.Close()
		return err
	}

	// ファイルを閉じる
	shootSound.Close()

	// デフォルトの音量とパンを設定
	soundManager.SetVolume("shoot", 0.7)
	soundManager.SetPan("shoot", 0.0)

	return nil
}
