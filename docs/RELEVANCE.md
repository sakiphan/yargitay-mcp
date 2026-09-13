# Hukuki uygunluk değerlendirmesi — 0.3.0

Kullanıcı tercihi: mevcut MCP istemci modeli değerlendirsin; ek AI API'si olmasın.
Sunucu arama ve temizlenmiş metin getirir; semantik puanlama yapmaz. İstemci zaten
aldığı metni değerlendirir. Bu özellik yeni sağlayıcı, anahtar veya AI maliyeti eklemez;
mevcut istemcinin veri işleme koşulları ve token kullanımı ayrıca geçerlidir.

## İstemciye verilen akış

1. Somut hukuki soru, maddi olay ve uygulanacak rejimi belirle.
2. Tek sayfadan en fazla üç adayla başla; otomatik sayfalama/tarama yapma.
3. Her adayın tam metnini gerekirse aynı hash/formatla parçalar halinde oku.
4. Aynı hukuki soru, uyumlu maddi olay/rejim, mahkemenin kendi gerekçesi, nihai
   hüküm/usul aşaması ve her sonuç için somut metin desteğini karşılaştır.
5. İç değerlendirmeyi uygun / uygun değil / belirsiz olarak ayır. Yalnızca uygun
   kararı, gerekçe–hüküm bağlantısını açıklayarak son yanıta al. Taraf iddiası,
   alt mahkeme kararı, karşı oy veya tetkik görüşünü mevcut hüküm sanma.
6. Uygun karar kalmadığında “İncelediğim adaylar arasında bu soruya uygunluğunu
   doğrulayabildiğim karar bulamadım” de. Erişim hatası veya eksik metin farklıdır;
   bunları uygunsuzluk ya da hiç karar olmadığı şeklinde sunma. Sayıyı tamamlamak
   için ilgisiz karar ekleme, filtreleri sessizce kaldırma.

Kullanıcı özellikle hata analizi veya ham arama adaylarını isterse elenen adaylar
uygunsuzluk durumu açıkça belirtilerek incelenebilir; emsal gibi sunulmaz.

## Sözleşme ve sınırlar

Dört kaynağın search/get yanıtlarında `relevance_review` bulunur. `status` daima
`not_assessed`, `evaluator` ise `calling_model` olur; bunlar sunucunun durumudur.
`presentation_rule=only_after_client_review` ve `required_checks` modelin görevini
hatırlatır. Aramada `relevance_checked=false` korunur. Sunucu modele değerlendirme
yaptıran ayrı bir servis çağırmaz; modelin değerlendirmesini geri alıp doğrulamaz.

`full_document_in_this_response=true` yalnızca bu yanıt offset 0'dan bütün metni
taşıdığında verilir. Son parça ve boş son offset tam metin sayılmaz. Model bütün
parçaları birleştirerek okumuş olabilir; bu alan okuma geçmişi tutmaz.

Danıştay `required_phrases` ana sorguyla AND olarak gönderilir. Resmî formun
`andKelimeler` dizisi kullanılır. Diğer kaynaklarda desteklenmeyen bu alan reddedilir.
Bu bir kelime filtresidir; terimler taraf anlatımında veya aktarılan mevzuatta
geçebilir. Hukuki uygunluk kanıtı değildir.

İstemci talimatları ve şemalar deterministik semantik filtre değildir. Model
talimatları yanlış uygulayabilir veya hukuki karşılaştırmada hata yapabilir.
Testler tüm modellerin yanıt kalitesini, güncelliği veya bağlayıcılığı kanıtlamaz.

## 12 Eylül 2026 kontrollü canlı gözlem

Tek geliştirme MCP sürecinde, üç tek-sayfa arama (page_size=1) ve iki belge
getirildi; ilk belgenin devam parçası aynı hash ile bellek önbelleğinden okundu.
Gerçek karar gövdeleri test fixture'ına, dosyaya veya uygulama loguna kaydedilmedi.

- `takdir komisyonuna sevk` + `zamanaşımı`: 1269 aday. İlk kararın 36.751 karakteri
  iki parçada okundu. Asıl uyuşmazlık engelli araç alımında ÖTV istisnasıydı;
  takdire sevkin zamanaşımına etkisi sorusu için bu incelemede elendi.
- `savunma hakkı` + `disiplin cezası` + `657`: 683 aday. İlk kararın 25.413 karakteri
  tam okundu. Öğrenci disiplin cezasının sicilden silinmesi hakkındaydı; olağan
  memur disiplininde savunma alınmaması sorusu için bu incelemede elendi.
- İlk sorguya anlamsız ek AND ifadesi konulduğunda sıfır sonuç döndü. Bu örnekte
  ek filtrenin yok sayılmadığını destekler; tüm sorgulara ilişkin garanti değildir.

Eleme bu oturumdaki istemci değerlendirmesidir, Go sunucusunun ürettiği bir karar
değildir. Bilinçli olarak ilgisiz iki adayı “uygun emsal” diye sunmadık. Bu sınırlı
örnekler pozitif emsal bulma başarısını veya üretimde hata oranını ölçmez.

Sentetik testler AND yükünü, normalizasyon/sınırları, desteklenmeyen alanların
ağdan önce reddini, dört kaynakta talimatları ve tam metnin otomatik hukuki onaya
dönüşmemesini kontrol eder. Gerçek istemci sohbet davranışı yeni binary ile MCP
yeniden başlatıldıktan sonra ayrıca denenmelidir.
