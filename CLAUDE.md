# CLAUDE.md

このファイルは、このリポジトリでコードを扱う際のClaude Code (claude.ai/code) 向けのガイダンスを提供します。

## コマンド

### ビルドと実行
- `go run main.go` - ポート5000でゲームサーバーを開始
- `go build -o gopher-war-II.exe main.go` - 実行可能ファイルをビルド
- 環境変数 `IP` でサーバーのIPアドレスを設定可能（例: `IP=0.0.0.0 go run main.go`）

### 開発
- サーバー起動後、`http://localhost:5000` でキャラクター選択画面にアクセス
- ゲーム画面は `/game?name={name}&character={character}` でアクセス
- サーバーは `./static/` ディレクトリから静的ファイルを提供
- WebSocketエンドポイントは `/ws` で利用可能

## アーキテクチャ

これはGoバックエンドとバニラJavaScriptフロントエンドで構築されたマルチプレイヤーシューティングゲームで、リアルタイム通信にWebSocketを使用しています。

### バックエンド (Go)
- **Webフレームワーク**: HTTPルーティングと静的ファイル提供にGin
- **WebSocket**: WebSocket接続とメッセージブロードキャストにMelodyライブラリ
- **ゲーム状態**: WebSocketセッションをキーとするマップを使用したインメモリストレージ
- **並行性**: スレッドセーフな操作のためのMutex保護共有状態
- **ゲームループ**: クライアント駆動のリフレッシュイベントがサーバーサイドの物理更新をトリガー

### データモデル
- `TargetInfo`: プレイヤーの位置、体力、チャージ状態、外観、移動状態管理
- `BulletInfo`: 弾丸の位置、移動、ダメージ、特殊タイプ（爆弾/ミサイル）
- `Config`: キャラクター別ゲームパラメータ（体力、速度、ダメージ、サイズ、移動制限）

### ゲームメカニクス
- プレイヤーはマウス移動で移動し、WebSocket経由で同期
- キャラクター別の移動速度制限（maxSpeed * speed）
- 2つの武器タイプ: 爆弾（クリック）とミサイル（ホールド＆リリース）
- 円形衝突検出（5pxマージン付き）でキャラクター別サイズ対応
- 初期位置保護：プレイヤーが移動するまで当たり判定無効
- サーバーが衝突検出、体力管理、死亡イベントを処理
- 20ms最小リフレッシュ間隔の50 FPSゲームループ

### フロントエンド (JavaScript)
- 3つのキャラクタータイプ（gopher、sushidog、duke）のキャラクター選択画面
- キャラクター別の設定値（configMap）による差別化
- プレイヤーの移動と武器方向のリアルタイムマウストラッキング
- 爆発と死亡パーティクルを含む視覚効果
- キャラクター固有のスプライト用の動的CSS更新
- ゲームオーバー後のキャラクター選択画面への復帰機能

### ファイル構造
- `main.go` - メインサーバー実装
- `static/index.html` - キャラクター選択画面
- `static/game.html` - ゲーム実行画面
- `static/{character}/` - キャラクター固有のスプライトアセット
- `original/` - ゲームのオリジナルバージョンを含む

### WebSocketプロトコル
ゲームはテキストベースのWebSocketメッセージを使用:

**クライアント→サーバー:**
- `init {name} {character} {config}` - プレイヤー初期化（キャラクター別config送信）
- `show {x} {y} {charge}` - プレイヤー位置/状態更新
- `fire-bomb/fire-missile {x} {y} {direction}` - 武器発射
- `refresh` - サーバーゲームループ更新のトリガー

**サーバー→クライアント:**
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