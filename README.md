# SearchAll: Search All CTF Challenges

CTF 問題のタグ検索ツールです。

## 機能

- 指定されたタグが含まれるチャレンジを検索
- 複数のタグを同時に検索可能

このツールには 2 つの機能が含まれます。

- **インタラクティブ検索モード**: 引数なしで実行すると、リアルタイムでタグをフィルタリング
- **静的検索モード**: 引数ありで実行すると、指定されたタグで即座に検索

## セットアップ

1. download dependencies

```bash
go mod download
```

2. build

```bash
go build -o searchall
```

## 使用方法

### インタラクティブモード（推奨）

```bash
# 引数なしで実行すると、リアルタイム検索モードに入ります
./searchall

# 静的検索モード
./searchall easy

# 複数のタグを指定して検索も可能
./searchall sql-injection geolocation
```
