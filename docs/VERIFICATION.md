# Doğrulama — 13 Eylül 2026

## 0.3.1 — sorudan arama planı

[Arama planlama sözleşmesi](QUERY_PLANNING.md), mevcut istemci modeline ayırt
edici ifade seçimi, kaynak yetenekleri, sabit filtreler, sınırlı alternatif ve
metinde doğrulanacak koşullar için talimat ve dört örnek verir. Ek API/model
çağrısı, backend sorgu yeniden yazımı veya yeni araç eklenmedi.

- Gofmt, `go vet ./...` ve çevrimdışı `go test -race ./...` geçti.
- Yeni test, dört kaynağın sekiz örnek çağrısını MCP şema doğrulamasından sentetik
  kaynaklarla geçirir; sabit filtreleri ve dokuz araç sınırını kontrol eder.
- npm sözdizimi kontrolü, 18 npm testi ve mevcut değerlendirme skorlayıcısının
  6 testi geçti. Bunlar istemci modelinin semantik başarısını ölçmez.
- Altı platform release derlemesi, npm paket içerik kontrolü ve altı binary'nin
  sürüm/hash/boyut doğrulaması geçti. Yalnız macOS arm64 çalıştırıldı.
- Gerçek paketlenmiş offline npx → Go → MCP testi initialize, dokuz araç,
  51 birim, temiz protokol stdout'u ve kapanışla geçti.
- Canlı mahkeme araması veya bağımsız model değerlendirmesi yapılmadı; önceki
  pilot sonuçları değiştirilmedi. İsabet artışı henüz ölçülmedi.
- Commit/push/publish, uzak CI ve kişisel MCP ayar değişikliği yapılmadı.
  ShellCheck/PowerShell bu turda yerelde çalıştırılmadı.

## Dosya yapısı refaktörü

[Mimari haritası](ARCHITECTURE.md) eklendi. `cmd` ince giriş noktasına dönüştürüldü;
CLI kodu `internal/cli` içine, kaynak arama/belge/JSON işlemleri ayrı dosyalara,
MCP talimatları ve görüntüleyici şablonu gömülü varlıklara ayrıldı. Araç veya API
davranışı değişikliği amaçlanmadı; sürüm 0.3.0 olarak kaldı, yayın yapılmadı.

- `make check`: gofmt, vet, tüm çevrimdışı Go/race testleri, npm check ve 18 npm testi geçti.
- `make build` ve altı platform release derlemesi geçti. Yalnızca macOS arm64 çalıştırıldı.
- Önceki ve yeni yerel binary'nin `initialize` yanıtı, dokuz araç şeması/açıklaması,
  CLI help/version çıktısı ve görüntüleyici HTML içeriği eşit bulundu. Binary'ler
  depo dışındaki bir çalışma dizininden açıldı; gömülü varlıklara erişim doğrulandı.
- Yeniden paketlenen npm arşivinin içerik/sürüm/hash doğrulaması ve gerçek offline
  npx → Go → MCP bağlantısı geçti. CI'a format kontrolünde `scripts` dizini eklendi;
  uzak CI çalıştırılmadı. ShellCheck/PowerShell yerelde çalıştırılmadı.
- Canlı mahkeme isteği, commit/push/publish ve kişisel MCP ayar değişikliği yapılmadı.

## Kapsam

0.3.0 istemci değerlendirme sözleşmesi ve Danıştay çoklu ifade daraltması
[RELEVANCE.md](RELEVANCE.md) içinde açıklanır. Semantik doğruluk garantisi değildir.
Bu sürüm için gofmt, `go vet ./...`, çevrimdışı `go test -race ./...`, 18 npm
testi, npm sözdizimi kontrolü ve `sh -n` geçti. Altı platform derlendi; paket
içeriği, sürüm ve altı binary'nin hash/boyut eşleşmesi doğrulandı. macOS arm64'te
paketlenmiş gerçek çevrimdışı npx → Go → MCP testi dokuz araç ve 51 birimle geçti.
Linux/Windows çalıştırılmadı; ShellCheck yerelde mevcut değil. Uzak CI/yayın yapılmadı.

0.2.0 sürümü Yargıtay'ın üç aracına AYM, Danıştay ve AİHM için altı araç ekler.
Yeni kaynakların kapsamı ve 0.2.1 Danıştay HTML düzeltmesi [COURTS.md](COURTS.md) içinde
belirtilir. Aşağıdaki eski sürüm gözlemleri kendi sürümlerine aittir. BAM eklenmedi.
Python projesi değiştirilmedi. AGPL-3.0-only tam metni ve önceki Apache-2.0
projesinin lisans/atıfları korundu. Git commit/push ve GitHub release yapılmadı.
Kullanıcının kişisel MCP ayarları bu geliştirme sırasında değiştirilmedi.

## Yerel kontroller

- Go 1.25.0, macOS arm64; resmî MCP Go SDK v1.7.0.
- `go mod verify`: başarılı.
- `gofmt` ve `go vet ./...`: başarılı.
- `go test -race -cover ./...`: offline testler, in-memory MCP, subprocess stdio,
  localhost HTTP ve görüntüleyici kontrolleri başarılı. Son kapsam değerleri aşağıda.
- Arama doğrulaması, metadata null değerleri, değişen upstream şeması, Unicode
  parçalar/hash, HTML temizleme, cache TTL/LRU/byte sınırı, eşzamanlı istek
  birleştirme, kuyruk iptali, retry ve erişim hataları sentetik testlidir.
- Kurulum: Codex/Claude Code/Gemini için geçici dizinlerde kayıt, idempotency,
  yedekleme, çakışmayı koruma, bozuk dosyayı koruma, dry-run ve kaldırma test edildi.
- POSIX indirme kurucusu sahte indirme aracıyla test edildi: doğru checksum ile
  kurulum, yanlış checksum ve indirme hatasında durma. Kişisel ayarlara dokunulmadı.
- macOS/Linux/Windows için amd64 ve arm64 binary'leri çapraz derlendi.
  macOS arm64 binary'nin `doctor` komutu çalıştırıldı.
- Release checksum listesi ve bağımlılıkların lisans bildirim dosyası üretildi.

## Doğrulanmayanlar

- Go 0.1.1 ifade araması için aşağıdaki sınırlı canlı kontrol yapıldı. Diğer
  filtre/sıralama/sayfalar ve sonuçların hukuki alaka düzeyi doğrulanmış değildir.
- Katalogdaki 51 seçenek Python projesindeki 9 Eylül gözleminden aynen taşındı.
  Katalog bugün yeniden indirilmedi; aktiflik iddia edilmez.
- Codex/Claude/Gemini kullanıcı arayüzlerinde gerçek kurulum denenmedi; MCP protokol
  ve config dosya testleri istemci UI uyumluluğunun tamamını kanıtlamaz.
- Linux ve Windows binary'leri bu macOS oturumunda çalıştırılmadı. Windows
  PowerShell kurucusu yerelde çalıştırılmadı; PowerShell burada kurulu değil.
- ShellCheck yerelde kurulu değil; yalnızca `sh -n` ve sahte indirme testleri çalıştı.
  CI'da ShellCheck ve PowerShell parse kontrolü tanımlandı; uzak CI çalıştırılmadı.
- Binary'lerde Apple notarization veya Windows Authenticode imzası yok.
- Genel internete HTTP sunumu, çoklu replica, toplu crawling ve otomatik sayfalama
  kapsam dışıdır. Her yerel süreç ayrı upstream bütçesine sahiptir.

## Yeniden çalıştırma

```sh
go mod download all
go mod verify
go vet ./...
GOPROXY=off go test -race -cover ./...
sh -n scripts/install.sh
sh scripts/release.sh v0.1.1
```

Testler için yalnızca localhost dinleme izni gerekebilir; Yargıtay ağı gerekmez.
Go bağımlılıklarını ilk kez indirmek ağ erişimi gerektirir; testler bunu yapmaz.

## Hedef dizindeki son kontrol

Yerel Go proje dizininde tüm paketler için
`go test -race -cover ./...` exit 0, `go vet ./...` exit 0. Gofmt farkı yok.
`sh -n` kurucu ve release scriptlerinde başarılı. Altı release binary'sinin
checksum doğrulaması başarılı. macOS arm64 binary `version`/`doctor` çalıştı.

| Paket | Statement coverage |
| --- | --- |
| cmd/yargitay-mcp | %44,5 |
| internal/server | %88,0 |
| internal/setup | %77,5 |
| internal/yargitay | %88,4 |
| scripts/notices | %77,6 |

POSIX kurucu testleri scripts paketindedir; shell satırları Go statement coverage
ölçümüne dahil değildir. Bu değerler branch veya hukukî doğruluk ölçümü değildir.
Local release binary'sinde commit `unknown`: henüz commit oluşturulmadı.

## npm başlatıcısı

- Node.js 22.14.0 / npm 11.6.0 ile yerel paket oluşturuldu. Haricî npm bağımlılığı
  veya install/postinstall hook'u yok. Son kullanıcı için Node.js >=20 gerekir.
- `npm run check` ve 15 npm testi başarılı; paket sürümü/hash uyumsuzluğu ve eksik
  release dosyalarında paketlemeyi reddetme de test edildi. `go vet ./...`, offline
  `go test -race ./...`, gofmt ve altı binary checksum kontrolü yeniden geçti.
- Altı platform seçimi, güvenilir HTTPS yönlendirmeleri, süre/boyut sınırları,
  SHA-256, bozuk cache onarımı, eşzamanlı ilk açılış, symlink reddi, argüman/sinyal
  aktarımı ve güvenli hata mesajları sentetik offline testlerle kontrol edildi.
- `npm pack` çıktısında yalnızca başlatıcı, sabit hash manifesti, README ve lisans/
  atıflar bulundu; native binary'ler ve kişisel ayarlar pakete dahil edilmedi.
- Paketlenmiş `.tgz`, gerçek `npx --offline` ile geçici npm/native önbelleğinde
  çalıştırıldı. Gerçek macOS arm64 Go binary'siyle initialize, tools/list ve
  list_yargitay_units çağrısı geçti: tam üç araç, 51 birim, temiz JSON-RPC stdout.
- Release scripti boş Git deposunda da commit `unknown` ile derleyebiliyor.
- GitHub/npm yayını, npm isim sahipliği ve yayımlanmış binary'nin gerçek ağdan ilk
  indirilmesi **doğrulanmadı**. Ağ indirmesi injected HTTPS yanıtlarıyla test edildi.
- Node 20/22 ve Linux/macOS/Windows npm test matrisi CI'a eklendi; burada yalnızca
  macOS/Node 22 çalıştırıldı. Uzak CI çalıştırılmadı.

## Otomatik npm yayın akışı

`npm-publish.yml`, yalnızca yayımlanmış kararlı GitHub release'inde çalışır.
Yayınlanan tarball ve altı binary indirilir; paket kimliği/sürümü, release sürümü,
manifest ve gerçek dosya hash/boyutları doğrulanır. Yeniden derleme yapılmaz.
`NPM_TOKEN` sadece publish adımında kullanılır; lifecycle scriptleri kapalıdır.
18 npm testi, mevcut gerçek release arşivinin yayın öncesi kontrolü ve paketlenmiş
offline npx–MCP testi başarılı. Üç workflow YAML sözdizimi kontrolünden geçti;
actionlint yerelde kurulu değil. GitHub Actions veya gerçek npm yayını yapılmadı;
secret'ın varlığı, token yetkisi ve npm paket adı sahipliği doğrulanmadı.

## 0.1.1 — ifade araması ve kısa yanıt düzeltmesi

12 Eylül 2026 canlı kullanımında düz `fazla çalışma` sorgusunun ilk sonucu adli
yardım hakkındaydı; ilk belgenin tamamında ifade yoktu. Aynı sorgu çift tırnakla
gönderilince farklı bir sonuç kümesi ve ifadeyi içeren ilk belge döndü. Bu gözlem,
tırnaksız sözdizimin beklenenden geniş arama yapabildiğini gösterir; upstream'in
iç arama algoritmasını veya bütün sorgularda aynı davranışı kanıtlamaz.

- İstemci artık düz metni normalize edip tırnaklı ifade olarak gönderir; önceden
  tırnaklanmış giriş çift kez sarılmaz. Gelişmiş sözdizimi kapsam dışıdır.
- Arama cevabına istenen normalize `filters`, `query_mode` ve metinlerin
  okunmadığını belirten `relevance_checked=false` eklendi. Otomatik belge indirme,
  sayfalama, daire taraması veya yeni MCP aracı eklenmedi.
- `probe --live`, sadece HTTP/JSON başarıyla geldi diye `search_verified=true`
  demez; ilk tam belgede örnek ifadenin varlığını da kontrol eder. İlgisiz veya boş
  sentetik sonuçlarda doğrulama işaretlenmediği regression testlerle kontrol edildi.
- Düzeltilen Go istemcisiyle **bir arama (page_size=1) ve bir belge** canlı probe'u
  başarılı: 7.500 karakter, `phrase_found_in_first_document=true`. Metin ve sorgu
  loglanmadı; gerçek karar metinleri fixture olarak kaydedilmedi. Ön teşhiste
  resmî ana sayfanın statik form/JS sözleşmesi sınırlı isteklerle incelendi ve
  mevcut MCP'de bir tırnaklı arama/ilk belge karşılaştırması yapıldı.
- MCP initialize talimatları ve araç açıklamaları kısa numaralı liste, okunan
  metinden kısa özet ve görüntüleme linki ister. HTML details/summary ve varsayılan
  uzun tam metin çıktısı engellenmesi yönünde talimat verilir. Bu, istemci modelinin
  her yanıtını zorunlu biçimlendiren bir renderer değildir; canlı sohbet yanıtı
  yeni binary ile istemci yeniden başlatılınca ayrıca denenmelidir.
- Gofmt, go vet, offline go test -race, 18 npm testi, altı platform build'i,
  tarball/hash doğrulaması ve gerçek offline npx–MCP bağlantısı kontrol edildi.
  GitHub/npm'e 0.1.1 yayını yapılmadı; Linux/Windows binary'leri yerelde çalıştırılmadı.
