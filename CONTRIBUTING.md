# Katkı

Go >=1.25. `AGENTS.md` kuralları geçerlidir. Testler sentetik ve offline tutulur.
Yeni şema davranışlarına regression testi ekleyin. Varsayılan hız sınırını azaltmayın;
CAPTCHA/erişim engeli aşma, toplu crawling ve otomatik sayfalama eklemeyin.

`gofmt`, `go vet ./...`, `go test -race ./...` ve binary derlemesi çalıştırılır.
Normal testler Yargıtay'a gitmez. Canlı probe rutin geliştirme/test döngüsü değildir.
Dağıtımda dokuz MCP aracının sözleşmesi ve Unicode offset/hash davranışı korunmalıdır.

## Yerleşim ve geliştirme

[Dosya haritası ve bağımlılık sınırları](docs/ARCHITECTURE.md) yeni kodun nereye
konacağını açıklar. `cmd` ince giriş noktasıdır; komut davranışı `internal/cli`
içinde test edilir. Testler kodun yanında tutulur; kaynak metinleri fixture yapılmaz.

İlk hazırlık: `go mod download all` ve `npm ci --ignore-scripts`. Sonrasında:

```sh
gofmt -l cmd internal scripts
go vet ./...
GOPROXY=off go test -race ./...
npm run check
npm test
go build -trimpath -o dist/yargitay-mcp ./cmd/yargitay-mcp
```

POSIX ortamında aynı kontroller için `make check`, yerel binary için `make build`
kullanılabilir. Windows'ta build çıkışını `dist/yargitay-mcp.exe` seçin. Canlı probe
bu hedeflere dahil değildir. Release/npm paketi için README'deki mevcut akış geçerlidir.

Katkılar AGPL-3.0-only altında sunulur; üçüncü taraf lisans/atıf koşulları korunur.
Başkasına ait kodu lisans kontrolü olmadan eklemeyin. Geçmişte Apache-2.0 ile dağıtılmış
kopyaların lisansı geriye dönük geri alınmış sayılmaz.
