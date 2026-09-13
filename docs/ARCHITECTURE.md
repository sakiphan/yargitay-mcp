# Proje yapısı

Bu depo tek Go modülü ve bağımlılıksız npm başlatıcısıdır. CLI, MCP protokolü,
kaynak adaptörleri ve yerel kurulum birbirinden ayrılır. Yeni bir servis, framework
veya çalışma zamanı dosya bağımlılığı eklenmez.

```text
cmd/yargitay-mcp/       Binary giriş noktası, linker sürüm bilgisi, stdio süreç testi
internal/
  cli/                 Komut yönlendirme ve uygulamanın başlatılması/kapatılması
  server/              MCP araçları ve yerel HTTP sunumu
    schema.go          Araç giriş şemaları
    instructions.go    Gömülü istemci talimatlarının yüklenmesi
    prompts/           Hukuki uygunluk ve yanıt talimatları
    http.go            Loopback MCP transport ve ortak HTTP güvenlik kontrolü
    viewer.go          Karar görüntüleyicisi
    templates/         Gömülü, HTML-escape uygulanan görüntüleyici şablonu
  courts/              AYM, Danıştay ve AİHM adaptörleri
    client.go          İstemci yaşam döngüsü ve ortak taşıyıcı bağımlılığı
    search.go          Arama isteği hazırlama ve yanıt ayrıştırma
    document.go        Belge isteği hazırlama ve temizleme
    decode.go          Ortak JSON alan/şema kontrolleri
    aym.go             AYM tür/metadata dönüşümü
    danistay.go        Danıştay HTML belge çıkarımı
    links.go           Yalnızca sabit resmî kaynak bağlantıları
  yargitay/            Yargıtay adaptörü ve ortak güvenli upstream altyapısı
  setup/               İstemci ayarlarına açık install/uninstall işlemleri
npm/
  bin/                 npx giriş noktası
  lib/                 Binary indirme, hash doğrulama ve çalıştırma
  scripts/             Paket hazırlama/doğrulama
  test/                Çevrimdışı başlatıcı testleri
scripts/               Yayın, OS kurucuları ve lisans bildirimi üretimi
docs/                  Tasarım, kaynak sözleşmeleri ve doğrulama notları
third_party/           Korunması gereken lisans ve atıflar
.github/workflows/     CI, release ve npm yayın akışı
dist/                  Üretilen binary/paketler; Git dışında
```

## Bağımlılık yönü

`cmd → cli → server / courts / setup / yargitay`

`server → courts / yargitay`, `courts → yargitay`.
Kaynak paketleri CLI veya MCP sunucusuna bağımlı değildir. `cli` bileşenleri kurar
ve kapatır; `main` yalnızca sürüm bilgisini geçirir ve süreç çıkışını yönetir.
Sürüm/commit `BuildInfo` ile aktarılır; release'in `main.version` ve `main.commit`
linker değişkenleri aynı kalır.

## Korunan sınırlar

- Tüm upstream HTTP I/O `internal/yargitay/client.go` içinde kalır. Kaynak adaptörleri
  bu taşıyıcı üzerinden çalışır; kendi `http.Client` nesnelerini oluşturmaz.
- Dört kaynak aynı süreç çapındaki sırayı ve hız sınırını kullanır. TLS, sabit
  origin, yönlendirme yasağı, sayfa sınırları ve otomatik tarama yasağı değişmez.
- `server/http.go` dış kaynağa erişmez; yalnızca yerel gelen HTTP isteklerini yönetir.
- `yargitay` paketinde ortak model/cache/limiter bulunması mevcut bilinçli sınırdır.
  Bu refaktör genel `utils`, `common`, `pkg` paketleri veya yapay katmanlar üretmez.
- Prompt ve HTML dosyaları `go:embed` ile binary'ye gömülür. Kurulumda ek dosya veya
  belirli bir çalışma dizini gerekmez. Bunları değiştirmek yeniden derleme gerektirir.
- Karar metni güvenilmeyen veridir; HTML şablonu `html/template` kullanır.
  Hukuki uygunluk talimatları renderer kodundan ayrıdır; sunucu hukuki onay üretmez.

## Dosya eklerken

Testi ilgili Go dosyasının yanında `*_test.go` olarak tutun. `cmd` testleri gerçek
stdio süreç girişini; `internal/cli` komutları ve sürüm aktarımını; `server` testleri
araç sözleşmesi, HTTP ve görüntüleyici güvenliğini kontrol eder. Sentetik fixture'lar
kullanılır; büyüyen fixture'lar ilgili paketin `testdata/` dizinine konabilir.

Yeni metadata/ayrıştırma davranışı `courts` veya `yargitay` içine; MCP şeması ve
istemci yönlendirmesi `server` içine; kurulum mantığı `setup` içine konur. HTTP
güvenlik katmanını bir kaynak eklemek amacıyla çoğaltmayın.

`Makefile` POSIX ortamlarında ortak geliştirme komutlarını sunar; zorunlu değildir.
Windows veya Make olmayan ortamlarda [CONTRIBUTING.md](../CONTRIBUTING.md) içindeki
doğrudan Go/npm komutları kullanılır. Üretilen `dist/`, npm manifesti ve bağımlılık
lisans çıktıları kaynak dosyalarla karıştırılmaz; release scriptleri bunları üretir.
