# Sorudan arama planı — 0.3.1

Önceki pilotta geniş konu ifadeleri ilk iki adayda somut soruyu karşılamadı.
Bu sürüm, aramadan önce **mevcut istemci modeline** plan hazırlama görevi verir.
Sunucuda yeni bir model, anahtar kelime çıkarıcı, semantik arama motoru veya AI
API çağrısı yoktur. Dokuz araç ve mevcut giriş/çıkış şemaları korunur.

## Planın parçaları

- Hukuki mesele: genel alan değil, çözülmesi istenen özel sorun.
- Kullanıcının verdiği olgular: statü, işlem, koşullar; eksikler uydurulmaz.
- Kaynak/tür/dil ve kullanıcı filtreleri: alternatifte de korunur.
- Ayırt edici ifade: kaynağa tek doğal ifade olarak gönderilir.
- Metinde doğrulanacaklar: şemada olmayan koşullar upstream filtresi sayılmaz.
- Gerekiyorsa tek alternatif: başarılı ama boş/alakasız aramadan sonra aynı kapsamda.

Model bu kısa planı hazırlayıp aracın desteklediği parametrelere dönüştürür. Plan
alanları MCP arama parametreleri değildir; `hukuki_mesele` gibi alanlar araca
gönderilmez. Kullanıcıya uzun düşünce zinciri veya ham plan JSON'u değil, kısa
arama kapsamı açıklaması verilir.

| Soru | İlk ifade örneği | Metinde ayrıca doğrulanacak |
| --- | --- | --- |
| Bordroyu aşan fazla çalışmanın tanıkla ispatı | `imzalı ücret bordroları` | İmza, ihtirazi kayıt, tahakkuk ve hükmün ispat yaklaşımı |
| Sevkin tarh zamanaşımına etkisi | `zamanaşımını durdurmak` + Danıştay AND: `takdir komisyonuna` | Sevkin amacı, tarh süresi, mahkemenin kendi gerekçesi |
| Kamu görevlisinin paylaşımı nedeniyle disiplin cezası | `sosyal medya paylaşımları nedeniyle` | Kamu görevlisi statüsü, disiplin yaptırımı, ifade özgürlüğü incelemesi |
| İlk polis sorgusunda avukat ve ifadenin kullanımı | ENG: `initial police questioning` | Avukat kısıtlaması, ifadenin mahkûmiyette kullanımı, hakkaniyet |

Bu ifadeler **davranışı gösteren örneklerdir**, resmî kaynakta her zaman geçen
kalıplar veya doğrulanmış en iyi sorgular değildir. Tırnaklı arama çekim/sözcük
sırasına duyarlı olabilir; aynı kavram farklı biçimde yazıldığında sonuç kaçabilir.
Örnekleri ezberlemek yerine yeni sorunun ayırt edici koşulunu bulmak hedeflenir.

## Kaynak sınırları

Yargıtay, AYM ve AİHM `query` alanını tek ifade olarak kullanır. Birbirinden kopuk
kelimeleri bu alana yazmak bağımsız AND koşulları oluşturmaz. Bu üç kaynakta
`required_phrases` desteklenmez. Danıştay'da en fazla beş ek ifade ile AND mümkündür.
Yeni desteklenmeyen filtreler, ham boolean sorgu veya wildcard eklenmez.

Kullanıcının açıkça istediği literal sorgu, metadata-only arama ve belirli belge
kimliğiyle getirme isteği korunur. Belirleyici hukuki rejim bilinmiyorsa modelden
önce açıklama istemesi beklenir; bilinmeyen statü/daire/tarih/kanun uydurulmaz.

## Bütçe ve durma

Varsayılan istemci planı: ilk sayfa, boyut 2; soru başına toplam en fazla 2 arama
çağrısı ve 3 farklı aday incelemesi. Kullanıcının daha düşük sınırı önceliklidir.
Bu sayılar istemci yönlendirmesidir, backend'de yeni bir oturum bütçesi sayacı yoktur;
sunucunun sayfa varsayılanı 10 ve mutlak üst sınırı 20 değişmedi.

Başarılı ama boş veya incelendiğinde açıkça ilgisiz adaylar döndüren ilk sorgudan
sonra aynı kapsamda tek bir alternatif mümkündür. Değişiklik kullanıcıya kısaca
açıklanır; sabit filtreler kaldırılmaz. İyi kanıt varsa ikinci arama yapılmaz.
Erişim/zaman aşımı/CAPTCHA/429/şema hatası, farklı sorgu veya kaynakla dolaşma
gerekçesi değildir. Otomatik sayfalama, toplu sorgu üretimi veya daire taraması yoktur.
Aynı kaynak+belge kimliği tekrar sayılmaz. Tam okuma tamamlanmadan emsal sunulmaz.

## Uygulama ve doğrulama

`internal/server/prompts/query_planning.txt` ve `query_examples.json` binary'ye
gömülür ve initialize talimatlarına eklenir. Ortak `search_review.txt` dört arama
aracında da kısa planlama kurallarını taşır; yalnız araç açıklamalarını kullanan
istemciler de temel yönlendirmeyi alır. Son hukuki uygunluk değerlendirmesi ayrı
kalır: sunucu `relevance_review.status=not_assessed` döndürmeye devam eder.

Yeni Go testi dört kaynağın sekiz örnek çağrısını gerçek MCP şema doğrulamasından,
sentetik kaynaklarla geçirir. Alternatiflerin yalnız ifadeyi değiştirdiğini,
kaynak/dil/tür/filtreleri koruduğunu; initialize sırasında örneklerin otomatik
çalışmadığını ve araç sayısının dokuz kaldığını kontrol eder. Bu, modelin her
soruda doğru plan üreteceğini veya hukuki uygunluğun arttığını kanıtlamaz.

Bu turda canlı mahkeme araması veya bağımsız model değerlendirmesi yapılmadı.
Önceki [pilot](EVAL_REPORT.md) değiştirilmeyen bir başlangıç gözlemi olarak kalır.
Sonraki karşılaştırma, eski ve yeni sorguları aynı önceden belirlenmiş aday/okuma
bütçesiyle ve mümkünse bağımsız uzman etiketleriyle ölçmelidir.

OpenAI Docs rehberindeki açık talimat ve farklı girdi/çıktı örnekleri yaklaşımı
uygulandı; belirli bir model seçilmedi veya API eklenmedi.
[Resmî prompt engineering rehberi](https://developers.openai.com/api/docs/guides/prompt-engineering#few-shot-learning).
