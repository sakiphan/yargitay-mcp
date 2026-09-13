# Sorudan arama planı

Kapsamlı arama değişikliği 0.3.2 ile gelir; 0.3.1 kullananların paket sürümünü
güncellemesi ve istemcinin MCP'yi yeniden başlatması gerekir. 0.3.1'deki iki aday/üç
inceleme ve ilk uygun sonuçta durma kuralları kaldırıldı.

Önceki pilotta geniş konu ifadeleri ilk iki adayda somut soruyu karşılamadı.
Bu sürüm, aramadan önce **mevcut istemci modeline** plan hazırlama görevi verir.
Sunucuda yeni bir model, anahtar kelime çıkarıcı, semantik arama motoru veya AI
API çağrısı yoktur. Dokuz araç ve mevcut giriş/çıkış şemaları korunur.

## Planın parçaları

- Amaç: konu taraması mı, somut hukuki soruya emsal araması mı?
- Hukuki mesele: kullanıcı somut soru vermişse çözülmesi istenen özel sorun.
- Kullanıcının verdiği olgular: statü, işlem, koşullar; eksikler uydurulmaz.
- Kaynak/tür/dil ve kullanıcı filtreleri: alternatifte de korunur.
- Ayırt edici ifade: kaynağa tek doğal ifade olarak gönderilir.
- Metinde doğrulanacaklar: şemada olmayan koşullar upstream filtresi sayılmaz.
- Kapsam: normal ilk sayfa, istenen sayı veya açık Yargıtay tümü/devam isteği.
- Alternatif: somut soruda başarılı ama boş/alakasız aramadan sonra aynı kapsamda;
  konu/özel ad için literal şartı yoksa açıklanan en fazla iki yazım.

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

## Kapsam ve durma

Normal başlangıç ilk sayfa ve boyut 10'dur; kullanıcının daha düşük sınırı korunur.
Bir uygun karar bulmak durma koşulu değildir. İstenen sayıya veya mevcut sayfadaki
adayların sonuna kadar inceleme yapılır; üç aday tavanı yoktur. Kullanıcı devam veya
tümü istemeden sonraki sayfalar kendiliğinden çağrılmaz; kalan kapsam açıklanır.

Yalnız Yargıtay için açık “tümünü / hepsini / ne kadar varsa” isteği, istemcinin
mevcut konu ve filtreler içinde sayfaları sırayla çağırmasına izin verir. İlk sayfa
boyutu 20 olabilir; sorgu, filtreler, sıralama ve sayfa boyutu o sorgu boyunca sabit
kalır. Her sayfanın yeni adayları okunur, sonra sonraki sayfa çağrılır. Backend'de
döngü, ek AI API'si veya otomatik crawler yoktur; her search yalnız bir sayfa getirir.
Bu istisna diğer kaynaklara veya konu dışı arşiv taramasına genişletilmez.

`total` kaynağın sorgu başına bildirdiği aday sayısıdır, uygun veya farklı sorgular
arasında benzersiz karar sayısı değildir. `null` bilinmiyor demektir. Tutarlı toplamın
son sayfası, başarılı boş sayfa veya toplam bilinmiyorken kısa son sayfa sorgunun
görülen sonudur. Değişen toplam, toplamla çelişen sayfa veya dolu sayfada hiç yeni
kimlik olmaması tutarsızlıktır: durulur ve kapsam eksik bildirilir. Şemanın 1000.
sayfa sınırında kalan varsa tamamlandı denmez; filtreleri bölerek sınır aşılmaz.

Erişim/zaman aşımı/CAPTCHA/429/şema hatasında durulur; alternatif sorgu/kaynakla
dolaşılmaz. Tek MCP süreci ve ortak kuyruk kullanılır: eşzamanlılık 1, varsayılan
en az 3 saniye aralık. Çok süreç açarak hız sınırı aşılmaz.

İlerleme sayıları ayrı tutulur: sorgu başına gözlenen toplam, benzersiz aday,
tam incelenen, uygun, ilgisiz, belirsiz/erişilemeyen ve bekleyen. Kaynak+kimlik
ile tekilleştirilir; varyant toplamları toplanıp benzersiz sayı diye sunulmaz.
Her sayfada konuşmada devam noktası tutulur: sorgu/filtre/sıralama/sayfa boyutu,
sonraki sayfa, incelenen/bekleyen kimlikler; yarım belge için offset/format/hash.
Süre/bağlam sınırı veya kullanıcı durdurması kısmi sonuç olarak raporlanır. Bu,
diskte kalıcı bir görev veya arka plan işi değildir.

Tam okuma tamamlanmadan uygun karar sunulmaz. “Seçili sorguların erişilebilen
sayfaları incelendi” denebilir; indeksleme, yazım, değişen sonuçlar ve kaynak
sınırları nedeniyle bütün Yargıtay arşivindeki ilgili kararların bulunduğu garanti
edilemez. Çok sonuç daha çok zaman ve mevcut istemci token kullanımı demektir.

## Konu ve yazım kapsamı

“Metin2 ile ilgili kararlar” gibi konu taramasında kullanıcının vermediği suç,
talep veya daireyle daraltma yapılmaz. Metinde oyunun uyuşmazlıkla gerçek ilişkisi
doğrulanır; ortak bir hukuki soruyu cevaplamayan kararlar sırf bu nedenle elenmez.
Konuya ilişkin karar ile somut davaya uygun emsal farklı etiketlerdir.

Literal ifade şartı yoksa, konu/özel ad için açıklanarak en fazla iki yazım ayrı
aranabilir: örneğin `Metin2` ve `Metin 2`. İlkinde ilgili karar bulunması ikinci
yazımı engellemez. Bunlar canlı doğrulanmış sonuçlar değil yazım örnekleridir.
Somut hukuki sorularda ise boş/açıkça alakasız başarılı aramadan sonra aynı kapsamda
tek alternatif kuralı korunur. Sınırsız eşanlamlı sorgu veya filtre gevşetme yoktur.

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
