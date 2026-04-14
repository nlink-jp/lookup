# lookup — アーキテクチャ

## 目的

JSONデータストリームを外部データソース（CSV/JSON）またはDNSで検索して
エンリッチするパイプ対応CLIツール。stdinからJSONを読み込み、設定された
ルールでフィールド値をルックアップテーブルと照合し、エンリッチされた
JSONをstdoutへ出力する。

## 動作モード

### データソースルックアップ（デフォルト）

```
stdin (JSON配列 or JSONL)
  → 各オブジェクトをパース
  → ルックアップフィールド値を抽出
  → 設定されたメソッドでデータソースと照合
  → マッチした行のフィールドをオブジェクトにマージ
  → エンリッチされたオブジェクトを出力
```

### DNSルックアップ（`--dns`）

```
stdin (JSON配列 or JSONL)
  → 各オブジェクトをパース
  → ルックアップフィールド値を抽出
  → IP vs ホスト名を判定
  → 逆引き(PTR) or 正引き(A)を実行
  → 結果をオブジェクトにマージ
  → エンリッチされたオブジェクトを出力
```

### 設定ファイル生成（`generate-config`）

```
データソースファイル (CSV/JSON/JSONL)
  → カラム名/キーを抽出
  → config.jsonテンプレートを生成
  → stdoutへ出力
```

## モジュール構成（新設計）

```
cmd/
  root.go          CLIエントリポイント、フラグ解析、モード分岐
config/
  config.go        Config構造体、Load(io.Reader)、Validate()
  mapping.go       マッピングルールパーサー
  path.go          データソースパス解決（~、相対パス）
match/
  matcher.go       Matcherインターフェース + ファクトリ
  exact.go         完全一致マッチング
  wildcard.go      グロブパターンマッチング (filepath.Match)
  regex.go         正規表現マッチング
  cidr.go          CIDRネットワークマッチング
source/
  loader.go        データソースインターフェース
  csv.go           CSVローダー
  json.go          JSON/JSONLローダー
dns/
  resolver.go      DNS正引き/逆引き、カスタムサーバー
process/
  enricher.go      コアエンリッチメントロジック (processObject)
  stream.go        入力フォーマット検出、JSONL/配列I/O
generate/
  generate.go      generate-configサブコマンド
main.go            全体の配線、cmd/を呼び出し
```

### 依存関係

```
main.go
  └── cmd/root.go
        ├── config/          (設定 + マッピング + パス)
        ├── source/          (CSV/JSON読み込み)
        ├── match/           (マッチングアルゴリズム)
        ├── dns/             (DNS解決)
        ├── process/         (エンリッチメント + I/O)
        └── generate/        (設定ファイル生成)
```

外部依存なし。標準ライブラリのみ。

## データフロー

### エンリッチメントパイプライン

```
reader ──► detectFormat()
              │
    ┌─────────┴──────────┐
    ▼                    ▼
  JSONL              JSON配列
    │                    │
    ▼                    ▼
 行ごと処理         json.Unmarshal
    │                    │
    └────────┬───────────┘
             ▼
     processObject(obj, mapping, lookupData, matcher, opts)
             │
     入力フィールド値を抽出
             │
      ┌──── DNSモード? ────┐
      ▼                   ▼
  findMatch()      dnsLookup()
      │                   │
      └────────┬──────────┘
               ▼
       OutputMap適用（フィールド選択 + リネーム）
               │
       元のオブジェクトにマージ
               │
               ▼
           writer
```

## マッチングメソッド

| メソッド | ルックアップフィールド | 入力値 | アルゴリズム |
|---------|---------------------|--------|-------------|
| `exact` | リテラル文字列 | リテラル文字列 | 文字列一致（デフォルト大文字小文字無視） |
| `wildcard` | グロブパターン | リテラル文字列 | `filepath.Match` |
| `regex` | 正規表現パターン | リテラル文字列 | `regexp.MatchString` |
| `cidr` | CIDR表記 | IPアドレス | `net.IPNet.Contains` |

CIDR以外の全メソッドが`case_sensitive`フラグをサポート（デフォルト: false）。

## 設定

### 設定ファイル（`-c`）

```json
{
  "data_source": "./users.csv",
  "matchers": [
    {
      "input_field": "user_lookup",
      "lookup_field": "username",
      "method": "exact",
      "case_sensitive": false
    }
  ]
}
```

### マッピングルール（`-m`）

```
<config_ref> as <input_field> [OUTPUT <src> [as <dst>], ...]
```

- `config_ref`: matcherの`input_field`を参照
- `input_field`: 入力JSONから値を取得するフィールド名
- `OUTPUT`: 任意のフィールド選択とリネーム
- OUTPUTなし: マッチした行の全フィールドが追加される

### データソースパス解決

1. `~/...` → ホームディレクトリ展開
2. 絶対パス → そのまま使用
3. 相対パス → 設定ファイルのディレクトリからの相対解決

## 入出力フォーマット

### 入力検出

最初の非空白バイトでフォーマットを判定:
- `[` → JSON配列（入力全体を1つの配列としてパース）
- それ以外 → JSONL（行ごとに処理）

### 出力フォーマット

- JSON配列入力 → 整形されたJSON配列出力（2スペースインデント）
- JSONL入力 → JSONL出力（コンパクト、1行1オブジェクト）

### マッチなしの動作

オブジェクトは変更なしでそのまま返される。フィールド追加なし、エラーなし。

## DNSモード

| 入力値 | ルックアップタイプ | 出力フィールド |
|--------|-------------------|---------------|
| 有効なIP | 逆引き(PTR) | `hostname` |
| IP以外 | 正引き(A) | `ip` |

カスタムサーバー: `--dns-server 8.8.8.8`（ポート未指定時は:53を付与）。

## エラー処理

| 条件 | 動作 |
|------|------|
| `-m`フラグ未指定 | 即時終了 |
| `-c`フラグ未指定（非DNSモード） | 即時終了 |
| 設定ファイル読み取り不可 | 即時終了 |
| マッチャー未検出 | 即時終了 |
| データソース読み取り不可 | 即時終了 |
| 不正なJSONL行 | 警告、行スキップ |
| マッチなし | 変更なしでパススルー |
| 入力フィールド未検出 | 変更なしでパススルー |
| 非文字列フィールド値 | 変更なしでパススルー |
| 正規表現コンパイルエラー | 警告、マッチなし |
| DNS解決失敗 | サイレント、エンリッチなし |

## テスト戦略

### ユニットテスト可能なモジュール

| モジュール | テスト対象 |
|-----------|-----------|
| `config/` | 設定パース、バリデーション、マッピングルール解析、パス解決 |
| `match/` | 各マッチャーの単体テスト（exact, wildcard, regex, CIDR） |
| `source/` | io.ReaderでのCSV/JSON/JSONL読み込み |
| `dns/` | モックnet.Resolverでのリゾルバ |
| `process/` | 依存注入によるエンリッチメントロジック |
| `generate/` | データソースからのキー抽出 |

### 結合テスト

ビルド済みバイナリ + testdataファイル — 既存のブラックボックステストカバレッジを維持。

### カバレッジ目標

| モジュール | 目標 |
|-----------|------|
| config/ | 90%+ |
| match/ | 95%+ |
| source/ | 90%+ |
| dns/ | 70%+（モックリゾルバ） |
| process/ | 85%+ |
| generate/ | 80%+ |
| **全体** | **80%+**（現在の2.4%から向上） |
