# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build and Run
- `go run main.go` - Start game server on port 5000
- `go build -o gopher-war-II.exe main.go` - Build executable
- Environment variable `IP` can set server IP address (e.g., `IP=0.0.0.0 go run main.go`)
- Environment variable `PORT` is automatically set in PaaS environments

### Development
- After starting server, access character selection at `http://localhost:5000`
- Game screen accessible at `/game?name={name}&character={character}`
- Server serves static files from `./static/` directory
- WebSocket endpoint available at `/ws`
- HTTPS environments automatically switch WebSocket to WSS (secure) connections

### Game Controls
- **Move**: Move cursor
- **Fire bomb**: Click
- **Fire missile**: Press and hold mouse button for 1+ seconds, then release

## Architecture

This is a multiplayer shooting game built with Go backend and vanilla JavaScript frontend, using WebSocket for real-time communication.

### Backend (Go)
- **Web Framework**: Gin for HTTP routing and static file serving
- **WebSocket**: Melody library for WebSocket connections and message broadcasting
- **Game State**: In-memory storage using maps keyed by WebSocket sessions
- **Concurrency**: Mutex-protected shared state for thread-safe operations
- **Game Loop**: Client-driven refresh events trigger server-side physics updates

### Data Models
- `TargetInfo`: Player position, health, charge state, appearance, movement state
- `BulletInfo`: Bullet position, movement, damage, special type (bomb/missile)
- `Config`: Character-specific game parameters (health, speed, damage, size, movement limits)

### Game Mechanics
- Players move via mouse movement, synchronized over WebSocket
- Character-specific movement speed limits (maxSpeed * speed multiplier)
- Two weapon types: bombs (click) and missiles (hold & release)
- Circular collision detection with 5px margin, character-specific sizes
- Initial position protection: no collision until player moves
- Server handles collision detection, health management, death events
- 50 FPS game loop with 20ms minimum refresh interval

### Frontend (JavaScript)
- Character selection screen with 3 character types (gopher, sushidog, duke)
- Character differentiation through configMap settings
- Real-time mouse tracking for player movement and weapon direction
- Visual effects including explosions and death particles
- Dynamic CSS updates for character-specific sprites
- Return to character selection after game over

### Key Files
- `main.go` - Main server implementation
- `static/index.html` - Character selection screen
- `static/game.html` - Game execution screen
- `static/{character}/` - Character-specific sprite assets

### WebSocket Protocol
The game uses text-based WebSocket messages:

**Client → Server:**
- `init {name} {character} {config}` - Player initialization with character config
- `show {x} {y} {charge}` - Player position/state update
- `fire-bomb/fire-missile {x} {y} {direction}` - Weapon firing
- `refresh` - Trigger server game loop update

**Server → Client:**
- `appear {id}` - Player appearance notification
- `show {id} {x} {y} {life} {name} {charge} {character}` - Player state broadcast
- `bullet {id} {x} {y} {direction} {special} {character}` - Bullet state broadcast
- `hit {target_id} {bullet_id} {life} {special} {character}` - Hit notification
- `miss {bullet_id} {special}` - Miss notification
- `dead {target_id} {bullet_id} {special}` - Death notification

### Character Configuration
Each character has unique settings for game balance:
- **gopher**: Fast but fragile, bomb-focused
- **sushidog**: Balanced, healing missiles
- **duke**: Powerful but slow, high damage

## Deployment

### Railway
- Direct Go binary deployment
- Automatic HTTPS with WebSocket auto-conversion to WSS
- Environment variable `PORT` automatically set
- Static files served at `/static/` path

### Known Issues and Solutions
- **Mixed Content Error**: HTTPS environments block `ws://` WebSocket connections
  - Solution: `static/game.html` uses `window.location.protocol` to auto-detect HTTP/HTTPS and select appropriate WebSocket protocol (ws/wss)
- **Environment Variables**: PaaS environments like Railway automatically set `PORT`
- **Static Files**: Gin static file serving configuration enables `/static/` path access

---

# 日本語版

このファイルは、このリポジトリでコードを扱う際のClaude Code (claude.ai/code) 向けのガイダンスを提供します。

## コマンド

### ビルドと実行
- `go run main.go` - ポート5000でゲームサーバーを開始
- `go build -o gopher-war-II.exe main.go` - 実行可能ファイルをビルド
- 環境変数 `IP` でサーバーのIPアドレスを設定可能（例: `IP=0.0.0.0 go run main.go`）
- 環境変数 `PORT` はPaaS環境で自動設定される

### 開発
- サーバー起動後、`http://localhost:5000` でキャラクター選択画面にアクセス
- ゲーム画面は `/game?name={name}&character={character}` でアクセス
- サーバーは `./static/` ディレクトリから静的ファイルを提供
- WebSocketエンドポイントは `/ws` で利用可能
- HTTPS環境では自動的にWebSocketもWSS（セキュア）接続に切り替わる

### ゲーム操作
- **移動**: カーソルを移動
- **爆弾発射**: クリック
- **ミサイル発射**: マウスボタンを1秒以上押し続けてからリリース

## アーキテクチャ

これはGoバックエンドとバニラJavaScriptフロントエンドで構築されたマルチプレイヤーシューティングゲームで、リアルタイム通信にWebSocketを使用しています。

### バックエンド (Go)
- **Webフレームワーク**: HTTPルーティングと静的ファイル提供にGin
- **WebSocket**: WebSocket接続とメッセージブロードキャストにMelodyライブラリ
- **ゲーム状態**: WebSocketセッションをキーとするマップを使用したインメモリストレージ
- **並行性**: スレッドセーフな操作のためのMutex保護共有状態
- **ゲームループ**: クライアント駆動のリフレッシュイベントがサーバーサイドの物理更新をトリガー

### データモデル
- `TargetInfo`: プレイヤーの位置、体力、チャージ状態、外観、移動状態
- `BulletInfo`: 弾丸の位置、移動、ダメージ、特殊タイプ（爆弾/ミサイル）
- `Config`: キャラクター別ゲームパラメータ（体力、速度、ダメージ、サイズ、移動制限）

### ゲームメカニクス
- プレイヤーはマウス移動で移動し、WebSocket経由で同期
- キャラクター別の移動速度制限（maxSpeed * speed倍率）
- 2つの武器タイプ: 爆弾（クリック）とミサイル（ホールド＆リリース）
- 円形衝突検出（5pxマージン付き）でキャラクター別サイズ対応
- 初期位置保護：プレイヤーが移動するまで当たり判定無効
- サーバーが衝突検出、体力管理、死亡イベントを処理
- 20ms最小リフレッシュ間隔の50 FPSゲームループ

### フロントエンド (JavaScript)
- 3つのキャラクタータイプ（gopher、sushidog、duke）のキャラクター選択画面
- configMapによるキャラクター別設定値での差別化
- プレイヤーの移動と武器方向のリアルタイムマウストラッキング
- 爆発と死亡パーティクルを含む視覚効果
- キャラクター固有のスプライト用の動的CSS更新
- ゲームオーバー後のキャラクター選択画面への復帰機能

### 主要ファイル
- `main.go` - メインサーバー実装
- `static/index.html` - キャラクター選択画面
- `static/game.html` - ゲーム実行画面
- `static/{character}/` - キャラクター固有のスプライトアセット

### WebSocketプロトコル
ゲームはテキストベースのWebSocketメッセージを使用:

**クライアント → サーバー:**
- `init {name} {character} {config}` - プレイヤー初期化（キャラクター別config送信）
- `show {x} {y} {charge}` - プレイヤー位置/状態更新
- `fire-bomb/fire-missile {x} {y} {direction}` - 武器発射
- `refresh` - サーバーゲームループ更新のトリガー

**サーバー → クライアント:**
- `appear {id}` - プレイヤー出現通知
- `show {id} {x} {y} {life} {name} {charge} {character}` - プレイヤー状態ブロードキャスト
- `bullet {id} {x} {y} {direction} {special} {character}` - 弾丸状態ブロードキャスト
- `hit {target_id} {bullet_id} {life} {special} {character}` - ヒット通知
- `miss {bullet_id} {special}` - ミス通知
- `dead {target_id} {bullet_id} {special}` - 死亡通知

### キャラクター設定
各キャラクターは独自の設定値を持ち、ゲームバランスが調整されています:
- **gopher**: 素早いが脆い、爆弾重視
- **sushidog**: バランス型、回復ミサイル
- **duke**: 強力だが遅い、高ダメージ

## デプロイメント

### Railway
- Go言語のバイナリを直接デプロイ可能
- 自動HTTPS対応（WebSocket接続も自動的にWSS接続に変換）
- 環境変数 `PORT` が自動設定される
- 静的ファイルは `/static/` パスで配信

### 既知の問題と対処法
- **Mixed Content Error**: HTTPS環境でWebSocket接続を `ws://` で行うとブロックされる
  - 対処法: `static/game.html` で `window.location.protocol` を使用してHTTP/HTTPS環境を自動判定し、適切なWebSocketプロトコル（ws/wss）を選択
- **環境変数**: Railway等のPaaS環境では `PORT` 環境変数が自動設定される
- **静的ファイル**: Ginの静的ファイル配信設定により `/static/` パスでアクセス可能