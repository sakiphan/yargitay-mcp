# Yargıtay MCP

Yargıtay, AYM, Danıştay ve AİHM/HUDOC karar kaynaklarını Codex, Claude Code/Desktop
ve Gemini CLI'ye bağlayan bağımsız, yerel Go uygulaması. BAM dahil değildir.
Mahkemeler veya model sağlayıcılarıyla resmî bağlantısı yoktur.

Geliştiriciler için: [Proje yapısı](docs/ARCHITECTURE.md) · [Katkı rehberi](CONTRIBUTING.md).

**Destek durumu:** Yargıtay ve Danıştay korunur; AYM ve AİHM desteği deneysel
olarak sunulur. Son sınırlı pilotta dört kaynakta arama/metin erişimi çalıştı;
AYM'de dar soruya uygun aday bulunamadı, AİHM'in uzun metinleri değerlendirme
bütçesini aştı. Bu etiket erişimin bozuk olduğu anlamına gelmez. Hiçbir kaynak
için hukuki uygunluk veya eksiksiz araştırma garantisi verilmez.
[Sorgu karşılaştırması ve ölçüm sınırları](docs/QUERY_EVAL_REPORT.md).

- Yerel stdio; karar istekleri kullanıcının bilgisayarından doğrudan resmî kaynağa gider.
- `npx` ile kurulumda yalnızca Node.js >=20 gerekir; Go/Python gerekmez.
- Doğrudan hazır binary kullananlar için Node.js de gerekmez.
- Dokuz araç, 51 paketlenmiş Yargıtay kurul/daire seçeneği, isteğe bağlı okunabilir karar sayfası.
- Bellek içi cache, 3 saniye ortak hız sınırı, tek sayfa, doğrulanmış TLS.
- **AGPL-3.0-only**. Ticari kullanım serbesttir; kapsanan türevler için copyleft yükümlülükleri geçerlidir.

## Kurulum

**Paket yayın durumu:** Kaynak koddan derleme desteklenir; hazır GitHub release
ve npm paketi henüz yayımlanmadı.
`npx` komutları hem GitHub `v0.3.1` release'i hem de `sakiphan-yargitay-mcp`
npm paketi yayımlandıktan sonra çalışır. Yayın öncesinde kaynaktan derleyin.

### Tek komutla MCP kaydı (npm)

Codex CLI:

```sh
codex mcp add yargitay -- npx -y sakiphan-yargitay-mcp
```

Claude Code:

```sh
claude mcp add --transport stdio yargitay -- npx -y sakiphan-yargitay-mcp
```

Gemini CLI / Claude Desktop için ilgili JSON ayarındaki `mcpServers` alanına
mevcut kayıtları koruyarak ekleyin:

```json
{
  "mcpServers": {
    "yargitay": {
      "command": "npx",
      "args": ["-y", "sakiphan-yargitay-mcp"]
    }
  }
}
```

Node.js ve `npx`, istemcinin PATH'inde bulunmalı; GUI uygulamalarında gerekirse
`npx` için mutlak yol kullanın. Kurulumdan sonra istemcinizi yeniden başlatın.
İlk çalıştırma binary indirir; istemcinin başlangıç süresi kısa ise önce terminalde
`npx -y sakiphan-yargitay-mcp doctor` çalıştırarak önbelleği hazırlayın.
Sürümü sabitlemek için paket adını `sakiphan-yargitay-mcp@0.3.1` yapın.
Görüntüleyici istenirse komutun sonuna `serve --viewer` ekleyin.

Başlatıcı macOS/Linux/Windows için x64 veya arm64 binary'sini **paketle aynı
sürümdeki** GitHub release'inden indirir. Boyutu ve npm paketine gömülü SHA-256
hash'ini doğrulamadan çalıştırmaz; önbellek her açılışta yeniden doğrulanır.
Hash, npm paketinin/release üretiminin güvenli olduğu varsayımına dayanır; bağımsız
kod imzası değildir. İndirme yalnızca izinli GitHub HTTPS adreslerini izler;
TLS doğrulaması kapatılmaz. İndirme süre sınırı 120 saniye, boyut sınırı 128 MiB.
Kurulum hook'u ve npm çalışma zamanı bağımlılığı yoktur. Başlatıcı MCP/stdout'u
değiştirmez, hata mesajlarını stderr'e yazar; kişisel istemci ayarlarını değiştirmez.

Binary önbelleği macOS'ta `~/Library/Caches/sakiphan-yargitay-mcp`, Linux'ta
`${XDG_CACHE_HOME:-~/.cache}/sakiphan-yargitay-mcp`, Windows'ta
`%LOCALAPPDATA%\sakiphan-yargitay-mcp\Cache` altındadır.
`YARGITAY_MCP_CACHE_DIR` ile mutlak özel yol seçilebilir. Bu önbellek karar değil,
yalnızca sürüm/hash'e göre ayrılmış executable dosyalarını saklar. İlk indirme
GitHub'a bağlanır; karar isteklerini yine Go çekirdeği resmî kaynağa gönderir.

### Node.js gerektirmeyen alternatif

Aşağıdaki kurucular yalnızca GitHub release'i yayımlandıktan sonra çalışır.

macOS / Linux — kurucuyu inceleyebilirsiniz: [scripts/install.sh](scripts/install.sh).
`codex` yerine `claude`, `claude-desktop` veya `gemini` seçilebilir:

```sh
curl --proto '=https' --tlsv1.2 -fsSL https://raw.githubusercontent.com/sakiphan/yargitay-mcp/v0.3.1/scripts/install.sh | sh -s -- --client codex
```

Windows PowerShell — [scripts/install.ps1](scripts/install.ps1):

```powershell
& ([scriptblock]::Create((Invoke-WebRequest https://raw.githubusercontent.com/sakiphan/yargitay-mcp/v0.3.1/scripts/install.ps1).Content)) -Client codex
```

Kurucu ilgili işletim sistemi/işlemci binary'sini ve SHA-256 listesini aynı release'den
indirir, hash'i doğrular, dosyayı kalıcı kullanıcı dizinine koyar ve **yalnızca seçilen
istemciye** kaydeder. Bu doğrulama release içeriğiyle tutarlılık kontrolüdür; bağımsız
bir imza/güven kaynağı değildir. Önce kurucu kodunu incelemek isteyenler dosyayı
indirip `sh install.sh --client codex` ile çalıştırabilir.

macOS/Linux hedefi `~/.local/bin/yargitay-mcp`, Windows hedefi
`%LOCALAPPDATA%\Programs\yargitay-mcp\yargitay-mcp.exe`. PATH'e ekleme gerekmez;
istemciye mutlak yol yazılır. Güncellemede eski binary `.previous` dosyasında korunur.

Kurulum sonrası istemcinizi yeniden başlatın. İstemcinin kendi kurulum/hesap
gereksinimleri ayrıca geçerlidir. `claude` Claude Code, `claude-desktop` masaüstü
uygulaması, `gemini` Gemini CLI'dir; web sohbetlerine yerel MCP kurulmaz.

### Kaynaktan derleme

Go >=1.25 gerekir; güncel desteklenen Go sürümü önerilir. Son kullanıcıya Go gerekmez.

```sh
git clone https://github.com/sakiphan/yargitay-mcp.git
cd yargitay-mcp
go build -trimpath -o yargitay-mcp ./cmd/yargitay-mcp
./yargitay-mcp doctor
./yargitay-mcp install --client codex --dry-run
./yargitay-mcp install --client codex
```

Windows'ta `-o yargitay-mcp.exe` ve `./yargitay-mcp.exe` kullanın. Binary'yi kurulumdan
sonra taşımayın; istemci kayıtları tam dosya yoluna bağlıdır.

### Ayarlar ve kaldırma

| İstemci | Varsayılan kullanıcı ayarı |
| --- | --- |
| Codex | `~/.codex/config.toml` |
| Claude Code | `~/.claude.json` |
| Claude Desktop / macOS | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Claude Desktop / Windows | `~/AppData/Roaming/Claude/claude_desktop_config.json` |
| Gemini CLI | `~/.gemini/settings.json` |

Kurucu dosyayı yedekler, diğer sunucuları korur, aynı kayıt varsa değişiklik yapmaz.
Başka bir `yargitay` kaydı, değiştirilmiş yönetilen ayarlar veya bozuk dosya varsa
üzerine yazmak yerine durur. Gemini JSONC yorumları desteklenmez; dosya değiştirilmez.
Codex TOML yorumları korunur. JSON dosyaları yeniden biçimlendirilir; diğer alanlar korunur.
Özel profil/config dizini kullananlar aşağıdaki elle kurulum yöntemini tercih etmelidir.

```sh
./yargitay-mcp uninstall --client codex
```

Bu yalnızca bu binary'yle eşleşen MCP kaydını kaldırır; binary ve yedekler korunur.

Elle Codex kaydı (binary yolu gerçek mutlak yol olmalı):

```sh
codex mcp add yargitay -- /tam/yol/yargitay-mcp serve --viewer
```

Diğer istemciler için stdio command: `/tam/yol/yargitay-mcp`, args:
`["serve", "--viewer"]`. Gemini `mcpServers` ayarı ve Claude Desktop JSON şeması
desteklenir. Resmî kaynaklar:
[Codex](https://learn.chatgpt.com/docs/extend/mcp?surface=cli),
[Claude Code](https://code.claude.com/docs/en/mcp),
[Gemini CLI](https://github.com/google-gemini/gemini-cli/blob/main/docs/tools/mcp-server.md).

## Araçlar

| Araç | İşlev |
| --- | --- |
| `search_yargitay_decisions` | Anahtar kelime, birim, esas/karar yıl-numara ve tarih aralıklarıyla tek sayfa arar |
| `get_yargitay_decision` | Sınırlı text/markdown parçası, resmî kaynak, hash ve devam offset'i döndürür |
| `list_yargitay_units` | 51 gözlenmiş form seçeneğini çevrimdışı listeler; `all/boards/civil/criminal` |
| `search_aym_decisions` / `get_aym_decision` | Bireysel başvuru veya norm denetimi araması ve metin |
| `search_danistay_decisions` / `get_danistay_decision` | Danıştay ifade araması ve resmî HTML görüntüleyicisinden temizlenmiş karar metni |
| `search_aihm_decisions` / `get_aihm_decision` | HUDOC hüküm/karar ve dil filtreli arama, metin |

Yeni araçların hepsinde `query` zorunludur; yalnızca tek kaynak/tek sayfa sorgulanır.
AYM: `decision_type=individual_application` (varsayılan) veya `norm_review`.
AİHM: `decision_type=judgments` (varsayılan) veya `decisions`; `language=ENG`
(varsayılan), `FRE` veya `TUR`. Türkçe çeviri kapsamı tam değildir; araç çeviri yapmaz.
Danıştay ifade, `required_phrases` ve sayfa/boyut parametrelerini destekler.
`required_phrases`, ana `query` ile birlikte aranacak en fazla beş ek düz metin
ifadedir (her biri 3–200, toplam normalize sorgu en fazla 1000 karakter).
Her ifade ayrı tırnaklanıp resmî `andKelimeler` dizisine gönderilir; aynı ifadeler
Türkçe büyük/küçük harf duyarsız birleştirilir. Bu AND filtresi yalnızca Danıştay'dadır.
Örnek: `{"query":"takdir komisyonuna sevk","required_phrases":["zamanaşımı"],"page_size":3}`.
Bu daraltma sözcükseldir; aynı kelimelerin bulunması aynı hukuki sorun demek değildir.
Desteklenmeyen filtreler sessizce yok sayılmaz, reddedilir. AYM sırası kaynağın
varsayılan alaka sırasıdır; Danıştay/AİHM tarihi azalan sırada istenir.

AYM örneği: `{"query":"ifade özgürlüğü","decision_type":"individual_application","page_size":3}`.
AİHM örneği: `{"query":"freedom of expression","language":"ENG","page_size":3}`.
Her `get_*` aracına **aynı kaynağın aramasından dönen** `id` verilir. AYM kimliği
`bb:UUID` veya `norm:UUID`, AİHM kimliği HUDOC biçimindedir; Yargıtay kimliğiyle
karıştırılmaz. Eksik metadata `null` kalır; başvuru numarası esas numarası yapılmaz.

**Danıştay 0.2.1 düzeltmesi:** Resmî belge yanıtı JSON değil HTML görüntüleyicisidir.
Karar `hiddencontent` alanındaki kodlanmış HTML'den temizlenerek alınır; sayfanın
menüsü, sorgu vurgusu ve scriptleri karar metnine eklenmez. Sayfada CAPTCHA scripti
bulunması erişim engeli sayılmaz. Canlı tek arama/tek belge testi başarılıdır;
bu, tüm kararların erişilebilir olduğunu garanti etmez. Gerçek erişim engelleri aşılmaz.
[Kaynak sözleşmeleri ve doğrulama kapsamı](docs/COURTS.md).

Yargıtay örnek araması:

```json
{"query":"fazla çalışma","unit":"9. Hukuk Dairesi","page_size":5}
```

En az bir filtre gerekir. `query` 3–1000 karakterlik düz metin ifadedir; upstream'e
otomatik çift tırnaklı ifade olarak gönderilir. Dıştaki düz/akıllı tırnaklar normalize
edilir; içeride çift tırnak veya ters eğik çizgi reddedilir. Boşluklar normalize
edilir. Gelişmiş sorgu sözdizimi ve semantik arama desteklenmez. İfadenin geçtiği
bir kararın asıl uyuşmazlığı farklı olabilir; metni okuyarak alakasını kontrol edin.
`filters` istenen normalize filtreleri, `query_mode` phrase/metadata_only/all_phrases ayrımını
gösterir. Arama metinleri indirmediği için `relevance_checked=false` döner; bu alanlar
kaynağın tüm filtreleri uyguladığına veya sonuçların hukuken ilgili olduğuna kanıt değildir.
Birim tam katalog adı olmalıdır; eksik metadata `null` kalır.
`case_year`, `decision_year`: 1900–2100; numara aralıkları: 1–99999999.
Tarihler `YYYY-MM-DD`; başlangıç bitişten büyük olamaz.
`sort_by`: `decision_date` (varsayılan), `case_number`, `decision_number`.
`sort_direction`: `desc` (varsayılan), `asc`. Sayfa 1–1000; boyut varsayılan 10,
maksimum 20. Otomatik sayfa/daire taraması yoktur.

Aramadan dönen kimlikle belge alın:

```json
{"document_id":"123456","output_format":"text","max_chars":30000}
```

Bu kimlik yalnızca sentetik örnektir. `next_offset` varsa onu `offset`, dönen
`content_sha256` değerini `expected_content_sha256` olarak gönderin. Format aynı kalsın.
Hash seçilen formatta normalize edilmiş **tam** metnin UTF-8 baytlarına aittir.
Offset/uzunluk Unicode kod noktası sayar, byte veya görsel karakter kümesi saymaz.
`max_chars` 1–100000; varsayılan 30000. Tam son offset boş son parça; ilerisi hatadır.
Markdown bağlantı/görselleri etkinleştirmeyen kaçışlanmış paragraf metnidir.

## İstemci yanıtı

**0.3.1: Sorudan arama planı.** Mevcut istemci modeli hukuki mesele ve kullanıcı
olgularını ayırır; genel konu kelimesi yerine ayırt edici doğal ifade seçer.
Desteklenmeyen koşulları filtre diye göndermek yerine karar metninde doğrular.
Başarılı ama boş/alakasız ilk aramadan sonra aynı kapsamda tek alternatif planlayabilir;
erişim hatası bunun gerekçesi değildir. Ek AI API'si veya otomatik backend sorgu
üretimi yoktur. [Planlama kuralları ve sınırları](docs/QUERY_PLANNING.md).

MCP talimatları kısa Türkçe, numaralı karar listesi, okunan ilk karar için 2–3
cümlelik özet ve **Kararı aç** bağlantısı ister. Varsayılan yanıtta tablo, HTML
`details`/`summary`, ham JSON veya uzun tam metin bulunmamalıdır. Kullanıcı açıkça
tam metni sohbete yazmayı isterse sade metin kullanılabilir.

**0.3.0: Değerlendirme mevcut istemci modelindedir; ek yapay zekâ API'si yoktur.**
Arama adayları doğrudan kullanıcıya emsal listesi olarak gösterilmemelidir. Model
tam metni okuyup somut hukuki soru, maddi olay/rejim, mahkemenin gerekçesi ve nihai
hükmü karşılaştırır. Son yanıtta yalnızca uygun bulduklarını, neden uygun olduğunu
metne dayandırarak sunar; belirsiz ve ilgisiz adaylarla liste doldurmaz. Uygun karar
yoksa bunu **incelenen adaylarla sınırlı** söyler; hiç karar bulunmadığını iddia etmez.

Dört kaynağın arama ve belge yanıtlarında `relevance_review.status=not_assessed`,
`evaluator=calling_model` ve beş değerlendirme kontrolü döner. Tam metnin teslimi
bile sunucu tarafından hukuki onay sayılmaz. `full_document_in_this_response`
yalnızca mevcut yanıtın kapsamıdır; son parça olması bütün metnin okunduğunu göstermez.
Bunlar istemci modeline verilen kurallardır; sunucuda semantik eleme veya modelin
uyacağını zorlayan bir mekanizma yoktur. Hatasızlık garantisi değildir.
[Değerlendirme akışı ve test sınırları](docs/RELEVANCE.md).

## Okunabilir karar görüntüleyicisi

`serve --viewer` veya `YARGITAY_VIEWER_ENABLED=true`, rastgele localhost portunda
oturuma özel adres açar. Kurucu bunu etkinleştirir. `view_url` varsa “Kararı aç”
bağlantısı olarak kullanılabilir. Sayfa yazdırılabilir; diske karar yazmaz.
Bağlantı yalnızca aynı bilgisayarda MCP çalışırken geçerlidir. Resmî ham kaynağı
gösteren `source_url` ayrıca korunur. Başka bilgisayara paylaşılabilir kalıcı link değildir.

## Sınırlar ve gizlilik

- Yalnızca dört sabit resmî HTTPS origin: `karararama.yargitay.gov.tr`,
  `kararlarbilgibankasi.anayasa.gov.tr`, `karararama.danistay.gov.tr`,
  `hudoc.echr.coe.int`; TLS açık, redirect/proxy kapalı.
- Bütün istemci nesneleri aynı süreç içindeki sırayı paylaşır: eşzamanlılık 1,
  başlangıçlar arasında en az 3 saniye. Ayrı Codex/Claude süreçleri veya farklı
  bilgisayarlar **ortak hız sınırı paylaşmaz**. Paralel toplu kullanım desteklenmez.
- Yargıtay'da 429/500/502/503/504 için en fazla 3 tekrar; ortak Retry-After/backoff beklemesi.
  Yeni kaynaklarda otomatik tekrar yoktur; 429/503 Retry-After ortak sırayı bekletir.
  Ağ hatası, timeout, 403 veya CAPTCHA otomatik yeniden denenmez/aşılmaz.
- İstek süresi varsayılan 45 saniye; kuyruk dahil işlem süresi 60 saniye.
  Açılmış yanıt sınırı 2 MB; kuyruk 20 çağrı.
- Her kaynak için ayrı bellek içi TTL/LRU: 256 öğe, 32 MB serileştirilmiş veri
  (dört kaynakta toplam en fazla 128 MB cache). Arama 300 saniye,
  belge 86400 saniye. Bu bir process RSS sınırı veya anonimleştirme değildir.
- Sorgu/metin loglanmaz; stdio stdout yalnızca protokoldür. Telemetri yoktur.
- MCP'nin yerel olması, kararın bağlanan yapay zekâ sağlayıcısına gönderilmediği
  anlamına gelmez: istemci araç sonuçlarını modeline aktarabilir.
- Bulunmuş bir kararın güncel/bağlayıcı olduğu veya araştırmanın tüm kararları
  kapsadığı iddia edilmez. Metni okumadan emsal sonucu çıkarılmamalıdır.

Resmî site sürümlü bir geliştirici API'si değildir. Otomatik erişim/yeniden kullanım
koşullarının tamamı doğrulanmış sayılmaz. [Gözlem ve test sınırları](docs/VERIFICATION.md).

## Yapılandırma

`.env` otomatik yüklenmez; değişkenleri sürece verin.

| Değişken (`YARGITAY_` önekli) | Varsayılan |
| --- | --- |
| `TIMEOUT_SECONDS` / `DEADLINE_SECONDS` | 45 / 60 |
| `MIN_INTERVAL_SECONDS` / `MAX_RETRIES` | 3 / 3 |
| `MAX_QUEUE` / `MAX_RESPONSE_BYTES` | 20 / 2000000 |
| `CACHE_ENABLED` | true |
| `SEARCH_CACHE_TTL_SECONDS` / `DOCUMENT_CACHE_TTL_SECONDS` | 300 / 86400 |
| `CACHE_MAX_ITEMS` / `CACHE_MAX_BYTES` | 256 / 32000000 |
| `MAX_PAGE_SIZE` | 20 |
| `VIEWER_ENABLED` | false; kurucu true ayarlar |

Yerel HTTP isteğe bağlıdır:

```sh
./yargitay-mcp serve --transport streamable-http --listen 127.0.0.1:8000
```

MCP `/mcp`, süreç healthcheck `/healthz`. Healthcheck upstream'e istek atmaz.
HTTP yalnızca loopback; Host kontrolü ve Origin reddi vardır. Genel internete servis
sunmak bu sürümün kapsamı değildir. Başka host'a bind desteklenmez.

## Geliştirme ve yayın

```sh
go mod download all
go mod verify
gofmt -w cmd internal scripts/notices
go vet ./...
GOPROXY=off go test -race -cover ./...
sh scripts/release.sh v0.3.1
npm ci --ignore-scripts
npm run check
npm test
npm pack --json --pack-destination dist > dist/npm-pack.json
npm run verify:pack
npm run test:pack
```

Normal testler sentetik transportlar kullanır; gerçek MCP transport testleri yalnızca
localhost/subprocess kullanır. Yargıtay paketinde unmocked ağ bağlantıları engellenir.
Canlı probe ayrıca açık istek gerektirir ve CI'da reddedilir:

```sh
./yargitay-mcp probe --live
./yargitay-mcp probe --live --source aym
./yargitay-mcp probe --live --source aym --decision-type norm_review
./yargitay-mcp probe --live --source aihm
```

Her komut yalnızca seçilen kaynakta en fazla bir küçük arama ve bir belge isteği yapar;
retry yoktur. `--source danistay` da desteklenir. Karar metni ve
sorgu çıktıya/dosyaya yazılmaz. `search_verified` ancak ilk belgenin tam metninde
örnek ifadenin varlığı kontrol edilirse true olur; bu, hukuki alaka doğrulaması değildir.
Güncel gözlem kapsamı [doğrulama notlarındadır](docs/VERIFICATION.md).

Release workflow, `v*` etiketi gönderildiğinde testlerden sonra **taslak** GitHub
release'i hazırlar. Paket sürümü ile etiket aynı olmalı.
Altı binary **bir kez** derlenir; `npm pack` aynı dosyalardan hash manifestini
üretir. Lisans/atıflar ve küçük `.tgz` npm paketi release'e eklenir; binary'ler npm
arşivine konmaz. Maintainer taslağı inceleyip GitHub'da **Publish release** dediğinde
`Publish npm` workflow'u otomatik olarak **o release'deki aynı tarball'ı** yayımlar.
Binary yeniden derlenmez. Etiket/paket/manifest sürümü, paket adı/lisansı/registry'si
ve altı binary'nin hash/boyutu tekrar doğrulanır; uyumsuzlukta npm yayını durur.
PR, normal branch push'u, taslak ve prerelease npm yayını yapmaz.

### Otomatik npm yayını için bir defalık ayar

1. npm'de paket oluşturma/yayımlama yetkili hesabından granular access token üretin:
   **Read and write (publish and stage)** ve etkileşimsiz yayın için **Bypass 2FA**.
   `stage only` token doğrudan yayın yapamaz. Erişimi mümkün olduğunca bu paketle
   sınırlayın; ilk yayın için yeni paket oluşturma yetkisi de gereklidir.
2. GitHub deposunda **Settings → Secrets and variables → Actions → New repository
   secret** yoluyla token'ı **`NPM_TOKEN`** adıyla ekleyin. Token'ı koda, `.npmrc`
   dosyasına veya sohbet mesajına koymayın; süresi dolunca secret'ı yenileyin.
3. Workflow'ları içeren kodu/etiketi gönderin; `Release` testleri taslağı hazırlasın.
   Taslak release'i GitHub arayüzünden yayımlayın. `Publish npm` işi otomatik başlar.

Token sadece son `npm publish` adımında `NODE_AUTH_TOKEN` olarak verilir; publish
`--ignore-scripts` ile çalışır. GitHub token'ının yetkisi publish işinde salt okumadır.
Secret eksik veya yetkisizse iş hata verir; GitHub release'i geri alınmaz. Secret'ı
düzelttikten sonra Actions üzerinden başarısız işi yeniden çalıştırabilirsiniz.
GitHub `GITHUB_TOKEN` ile üretilen olaylar yeni workflow'ları tetiklemeyebilir;
bu yüzden mevcut akış taslak release'in insan tarafından yayımlanmasını bekler.

Kaynaklar: [GitHub npm yayın akışı](https://docs.github.com/en/actions/tutorials/publish-packages/publish-nodejs-packages),
[npm token yetkileri](https://docs.npmjs.com/creating-and-viewing-access-tokens/).
Paket adı uygunluğu veya npm hesap yetkisi yerelde doğrulanmadı. npm sürümü
değişmez olmalı: yayımlanmış binary'leri yeniden derleyip değiştirmeyin; güncelleme
için yeni sürüm/etiket kullanın. Kaynak, LICENSE, NOTICE ve THIRD_PARTY_NOTICES
binary'lerle birlikte erişilebilir tutulmalıdır. `test:pack` macOS/Linux'ta geçici
önbellekle, ağsız gerçek `npx` → Go → MCP testi yapar; kişisel ayarları değiştirmez.

## Lisans

[GNU AGPL-3.0-only](LICENSE). Değiştirilmiş sürüm dağıtıldığında kapsanan kaynak kodu
alıcılarına sunulur; değiştirilmiş ağ hizmetinde AGPL §13 uyarınca kullanıcılarına
kaynak erişimi sağlanır. Ticari kullanım yasak değildir. Bağımsız bir uygulamanın
MCP'yi çağırması o uygulamanın bütün kodunu otomatik olarak bu lisansa sokmaz.
Bu özetin yerine lisansın tam metni geçerlidir. Karar metinleri yazılım lisansına
tabi tutulmaz. Ayrıntılar: [NOTICE](NOTICE), [SECURITY](SECURITY.md).
