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

## モジュール構成

```
main.go              CLIフラグ、execute()、モード分岐
config.go            Config/Matcher/Mapping構造体、LoadConfig()、ParseMapping()
match.go             FindMatch()、matchExact/Wildcard/Regex/CIDR
source.go            LookupData型、LoadCSV()、LoadJSON()、LoadSource()
dns.go               dnsResolverインターフェース、dnsLookup()、newResolver()
process.go           enrichObject()、processStream()（JSONL/配列自動検出）
generate.go          generateConfig()、extractCSVHeaders()、extractJSONKeys()
path.go              ResolveDataSourcePath()
```

### 依存関係

```
main.go（CLIシェル）
  └── execute()
        ├── config.go   (LoadConfig, ParseMapping, FindMatcher)
        ├── path.go     (ResolveDataSourcePath)
        ├── source.go   (LoadSource)
        ├── match.go    (FindMatch)
        ├── dns.go      (dnsLookup, newResolver)
        ├── process.go  (enrichObject, processStream)
        └── generate.go (generateConfig)
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
     enrichObject(obj, mapping, data, matcher, dnsMode, resolver)
             │
     入力フィールド値を抽出
             │
      ┌──── DNSモード? ────┐
      ▼                   ▼
  FindMatch()      dnsLookup()
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

### ユニットテスト

| テストファイル | カバレッジ |
|--------------|-----------|
| `config_test.go` | Config解析、FindMatcher、ParseMapping（100%） |
| `match_test.go` | exact/wildcard/regex/cidr、大文字小文字、エッジケース（コア100%） |
| `source_test.go` | io.ReaderでのCSV/JSON読み込み（93%+） |
| `dns_test.go` | モックリゾルバ: 正引き/逆引き、成功/失敗（100%） |
| `process_test.go` | enrichObject、processStream JSONL/配列、不正入力 |
| `generate_test.go` | CSV/JSONキー抽出、エラーケース（90%+） |
| `path_test.go` | パス解決: チルダ、絶対、相対（100%） |

### 回帰テスト

`main_test.go` が `execute()` 経由で既存 `testdata/` ファイルに対してテスト:
- exact match（OUTPUTフィールドマッピング付き）
- wildcard match（全フィールド出力）
- regex match
- CIDR match（JSON配列入出力）

### カバレッジ

| 指標 | 値 |
|------|---|
| 全体 | 77%+ |
| コアロジック（config, match, dns, path） | 95%+ |
| main() / handleGenerateConfigCmd() | 0%（薄いCLIシェル） |
