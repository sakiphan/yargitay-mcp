# Ek karar kaynakları — 0.2.0 / 0.2.1 / 0.3.0

Gözlem tarihi: 12 Eylül 2026. Bunlar resmî web arayüzlerinin gözlenen istekleridir;
sürümlü, kararlılık garantili geliştirici API'leri olarak sunulmaz.

## Kaynaklar ve araç sözleşmesi

- AYM: https://kararlarbilgibankasi.anayasa.gov.tr/kbb/
  `POST /api/core/public/search`. `query`, `kararTipi`, `page`, `size`;
  belge için aynı uç noktaya `id`, `kararTipi`, `size=1`. `data[].icerik` temizlenir.
  Bireysel başvuru ve norm denetimi ayrı sorgulanır. UUID'ye `bb:`/`norm:` öneki
  eklenerek tür karışıklığı önlenir. Resmî görüntüleme bağlantısı arayüzün kullandığı
  URL-safe base64 `kbb:UUID` kodlamasını kullanır; bu bir erişim yetkisi değildir.
  İndirme/CAPTCHA doğrulama uç noktaları kullanılmaz; doğrulama başlığı üretilmez.
- Danıştay: https://karararama.danistay.gov.tr/
  `POST /aramalist`, `data.andKelimeler` içinde ayrı çift tırnaklı ifadeler;
  0.3.0'da ana `query` yanında en fazla beş `required_phrases` AND koşulu eklenebilir.
  `pageNumber`, `pageSize`, `siralama=3`, `siralamaDirection=desc`.
  Metadata `daireKurul`, `esasNo`, `kararNo`, `kararTarihi`; sayı `recordsFiltered`.
  Resmî form belgeyi `GET /getDokuman?id=...&arananKelime=...` üzerinden açar.
  0.2.0'da HTML yanıtı yanlışlıkla erişim hatası sayılıyordu; 0.2.1 bunu düzeltir.
  Belge HTTP 200 `text/html;charset=UTF-8` döndürür. Görüntüleyici
  `#hiddencontent.text()` içindeki HTML'yi görünür alana yerleştirir. Adaptör yalnızca
  bu alanı okur, entity kodlamasını bir kez çözer ve karar HTML'sini temizler.
  JSON hata zarfı metin sayılmaz; genel uygulama hatası erişim engeli diye sunulmaz.
- AİHM: https://hudoc.echr.coe.int/eng
  `GET /app/query/results`; `contentsitename=ECHR`, `documentcollectionid2`,
  `languageisocode` ve çift tırnaklı düz metin. `start` sayfa offset'i, `length` boyut.
  `GET /app/conversion/docx/html/body?library=ECHR&id=...` ile HTML metin alınır.
  `JUDGMENTS`/`DECISIONS` koleksiyonu açıkça seçilir; dil değiştirilmez.
  Arama yanıtındaki koleksiyon ve dil istenen filtrelerle eşleşmezse hata verilir.
  Dil/çeviri statüsü ve yayımlanmış belge kapsamı hukuki değerlendirme yerine geçmez.

Yalnızca resmî açık arayüzlerden yararlanılır; oturum açma, CAPTCHA çözme, proxy/TLS
gevşetme veya kaynak değiştirerek erişim engelini aşma uygulanmaz. Otomatik arama
yayılımı, sayfalama ve toplu indirme yoktur. Statik JS arayüzleri sözleşme keşfinde
okundu; depo içine kopyalanmadı. Gerçek karar gövdeleri fixture, dosya veya loga yazılmadı.

## Güvenlik ve çalışma biçimi

Tüm HTTP giriş/çıkışı `internal/yargitay/client.go` sınırında kalır; kaynak adaptörleri
`internal/courts` içindedir. Dört kaynak aynı süreç çapında FIFO sırayı paylaşır:
eşzamanlılık 1, başlangıçlar arasında en az 3 saniye. Her kaynağın cache'i ayrıdır.
Yeni adaptörlerde otomatik yeniden deneme yoktur. Süre/boyut/sayfa/kuyruk sınırları
mevcut `YARGITAY_*` ayarlarından alınır. Hata veya bozuk şema başarı olarak cache'lenmez.

Altı yeni araç mevcut üç araca eklenir; kurulum/paket adı `sakiphan-yargitay-mcp`
olarak korunur. `serve --viewer` her kaynağı ayrı yolda, aynı rastgele token/loopback
güvenlik kontrolleri altında açar. Metinler HTML olarak yürütülmez; kaynak adı gösterilir.
Belge parçalama/hash/Unicode offset davranışı Yargıtay ile aynıdır.

## Doğrulama sınırı

Canlı keşifte AYM bireysel başvuru ve norm denetimi için birer arama sonucu ve
`icerik` alanı; AİHM'de bir arama sonucu ve HTML metni görüldü. Danıştay arama
yanıtı görüldü; 0.2.0'da belge doğrulanamadı, aşağıdaki 0.2.1 testi bunu tamamladı.
Bu gözlemler tüm tarihleri, dilleri,
karar türlerini veya kullanım koşullarının tamamını doğrulamaz.

Normal testler sentetik taşıyıcılar kullanır; canlı probe CI'da engellenir.
`probe --live --source ...` yalnızca seçilen kaynakta bir arama ve en fazla bir
belge isteği yapar. `search_verified` yalnızca örnek ifadenin ilk metinde bulunmasıdır,
hukuki alaka veya bağlayıcılık doğrulaması değildir. Danıştay hatası saklanmaz.

Son uçtan uca probe sonuçları: AYM bireysel başvuru 22.941, AYM norm denetimi
18.986, AİHM 50.608 Unicode kod noktası uzunluğunda temizlenmiş metin döndürdü;
örnek ifade üçünün ilk metninde de bulundu. Gerçek karar gövdeleri saklanmadı.
Danıştay'ın ilk hata sınıflandırması bir erişim engelini kanıtlamıyordu. Engel aşma
veya otomatik yeniden deneme uygulanmadı; sonraki kontrollü teşhis aşağıdadır.

Yerel doğrulama: gofmt, `go vet ./...`, çevrimdışı `go test -race ./...`, 18 npm
testi ve npm sözdizimi kontrolleri geçti. Altı platform binary'si çapraz derlendi;
Linux/Windows binary'leri yerelde çalıştırılmadı. Paket içeriği ve yayın hash/sürüm
eşleşmesi doğrulandı. Gerçek çevrimdışı npx → Go → MCP testi dokuz aracı ve 51
Yargıtay birimini kontrol eder. ShellCheck yerelde kurulu olmadığı için çalıştırılamadı;
CI'daki kontrol korunur. Bu çalışma commit/push/publish veya kişisel MCP ayar değişikliği yapmaz.

## 0.2.1 — Danıştay belge teşhisi ve düzeltmesi

Kullanıcının açık isteğiyle kontrollü teşhis yapıldı. Yanıt HTTP 200 ve HTML idi;
sayfadaki bir CAPTCHA scripti yüzünden eski genel kontrol `ErrAccess` üretiyordu.
Görünür `content` alanı boş, `hiddencontent` alanı entity-kodlanmış karar HTML'siydi.
Kod dış sayfayı metin saymak yerine yalnızca bu alanı ayrıştırır. Örnek görünümde
sayfanın tüm temiz metni 2.088 karakterken gerçek karar metni 1.652 karakterdir.

Tamamlanan adaptörle `probe --live --source danistay` bir arama ve bir belge
çağrısında başarılı oldu: `document_verified=true`, `document_chars=1652`,
`phrase_found_in_first_document=true`. Bu, hukuki alaka doğrulaması değildir.
Boş `arananKelime` parametresiyle de çalıştığı gözlendi; parametre atılmaz ancak
isteğe sorgu vurgusu eklenmez. Cookie jar veya yönlendirme takibi gerekmedi.

Gerçek 401/403 erişim reddidir; 3xx izlenmeyen yönlendirme olarak ayrılır.
`FMTY=ERROR` genel uygulama hatasıdır, otomatik olarak erişim reddi değildir.
Yalnızca bir script adında `captcha` geçmesi engel sayılmaz. Karar alanı olmayan
gerçek doğrulama widget'ı reddedilir; bilinmeyen/boş/çift karar alanları şema hatasıdır.
Ham karar, sorgu, çerez veya yanıt gövdesi loglanmaz; teşhis geçici kodu dağıtıma eklenmez.

Sentetik regresyon testleri: entity çözme, paragraf/Unicode koruma, script ve dış
sayfa ayıklama, yanlış CAPTCHA tespiti, gerçek widget, eksik/çift alan, JSON runtime
hatası, HTML/JSON içerik türleri ve yönlendirme/erişim hatası ayrımı.
