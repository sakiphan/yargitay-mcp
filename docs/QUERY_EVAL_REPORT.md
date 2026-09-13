# Sorgu karşılaştırması — 13 Eylül 2026

**Bu küçük pilotta doğrulanan uygun aday sayısı eski sorgularda 0/8, yeni
sorgularda 2/8 oldu. Artış yalnız Danıştay sorusunda görüldü.** Bu, genel hukuki
doğruluk oranı veya yeni promptun bütün istemcilerde başarı kanıtı değildir.

## Yöntem

Önceki pilotun dört sorusu kullanıldı. Eski sorgular ve 0.3.1 ile gönderilen
örneklerin ilk sorguları aynı oturumda, aynı bağlı MCP üzerinden karşılaştırıldı.
Her yöntem/kaynak için bir arama, ilk sayfa ve iki aday: toplam sekiz arama,
16 farklı aday. Alternatif sorgu, sayfalama, kaynak taraması veya değerlendirici
tarafından hata tekrarı yapılmadı. Kaynaklar arasındaki sıra eski/yeni, yeni/eski
şeklinde dönüşümlüydü; rastgele veya kör değildi.

[Plan](../evals/relevance/runs/2026-09-13-query-plan.json) canlı aramalar başlamadan
kaydedildi. Aynı daire, belge türü ve dil filtreleri korundu; yalnız arama ifadeleri
değişti (Danıştay'da `query` ile ek AND ifadesinin ikisi). Sıralama iki yöntemde
aynı kaynak varsayılanında bırakıldı. Sonuç toplamları aşağıda kaynakça bildirilen
sözcüksel arama büyüklükleridir, hukuki uygunluk sayıları değildir.

| Kaynak | Eski sorgu | Yeni sorgu | Toplam eşleşme eski → yeni |
| --- | --- | --- | ---: |
| Yargıtay | fazla çalışma | imzalı ücret bordroları | 56.242 → 3.833 |
| Danıştay | takdir komisyonuna sevk AND zamanaşımı | zamanaşımını durdurmak AND takdir komisyonuna | 1.269 → 508 |
| AYM | sosyal medya | sosyal medya paylaşımları nedeniyle | 579 → 42 |
| AİHM | access to a lawyer | initial police questioning | 545 → 11 |

Belge başına önceden belirlenen üst sınır **60.000 karakter**, parça boyutu
20.000 karakter ve en fazla üç get çağrısıydı. Sınırı aşan belge başka kısa bir
adayla değiştirilmedi. Sınır içindeki bütün metinler başlangıçtan sona, aynı
format/hash ve kesintisiz offset ile okundu. AİHM belgelerinin ilk parçaları araçtan
geldi fakat boyutları görüldükten sonra modele gösterilmedi; hukuki etiket verilmedi.
Bu bir istemci okuma bütçesidir; sunucu belgeyi upstream'den zaten indirmiş olabilir.

Toplam 21 belge/parça çağrısı yapıldı; 12 belgenin 212.304 karakteri tam okundu.
Bu turda sekiz aramanın ve 16 farklı belgenin erişiminde araç hatası görülmedi.
Önceki gündeki zaman aşımları bu koşuya taşınmadı.

## Sonuçlar

| Kaynak | Eski: uygun / aday | Yeni: uygun / aday | İnceleme sonucu |
| --- | ---: | ---: | --- |
| Yargıtay | 0/2 | 0/2 | İki yöntemde de ikişer belirsiz |
| Danıştay | 0/2 | 2/2 | Eski iki aday ilgisiz; yeni iki aday uygun |
| AYM | 0/2 | 0/2 | İki yöntemde de ikişer ilgisiz |
| AİHM | Değerlendirilemedi | Değerlendirilemedi | Her yöntemde iki belge okuma sınırını aştı |

Her yöntemin sekiz adayı paydada tutulduğunda **doğrulanabilen uygun aday verimi
%0 → %25 (+25 yüzde puan)**. Yalnız tam okunanlar için 0/6 → 2/6; iki yöntemde
de okuma kapsamı 6/8. Sınır nedeniyle değerlendirilemeyen adaylar uygunsuz değildir.
Bir soruda en az bir uygun karar bulma gözlemi 0/4 → 1/4; AİHM'in hukuki sonucu
bilinmiyor. Bağımsız gold olmadığı için doğrulanmış precision/recall verilmez.

### Uygun bulunan iki aday

1. **Danıştay 7. Daire, 2024/3528 E., 2025/3711 K., 18.11.2025.**
   Çoğunluk, 2017 ÖTV tarhiyatında zamanaşımını durdurma saikiyle yapılan sevki
   somut tarihlerle inceliyor; geç tebliğ edilen tarhiyat nedeniyle kararı bozuyor.
   Kanun aktarımı veya taraf iddiası değil, “Hukuki Değerlendirme” ve “Karar Sonucu”
   birlikte destekliyor. Karşı oy, çoğunluğun sonucu olarak sunulmadı.
   [Resmî metin](https://karararama.danistay.gov.tr/getDokuman?arananKelime=&id=1219577400).
2. **Danıştay 7. Daire, 2024/3029 E., 2025/3662 K., 17.11.2025.**
   Ödeme emri davasında, 2014 BSMV tarhiyatının takdire sevki ve geç tebliğinin
   tarh zamanaşımına etkisi inceleniyor. 2014 kısmı bozuluyor; 2015 kısmı onanıyor.
   Dava tahsil aşamasında olsa da ilgili çoğunluk gerekçesi doğrudan tarh
   zamanaşımına ilişkin. İki dönem ve kısmi sonuç birbirine karıştırılmadı.
   [Resmî metin](https://karararama.danistay.gov.tr/getDokuman?arananKelime=&id=1219576200).

### İyileşmeyen alanlar

- **Yargıtay:** Yeni ilk adayda bordro ve yazılı ispat kuralı gerçekten mahkemenin
  gerekçesinde bulunuyor; ancak somut bozma kaloriferci/kantin çalışmasının süre
  hesabına ilişkin. İmzalı, ihtirazi kayıtsız, fazla çalışma tahakkuklu bordroyu
  aşan miktarın tanıkla ispatının somut uyuşmazlıkta uygulandığı doğrulanmadı.
  İkinci yeni adayda bordro tahakkuku genel tatil ücretine ilişkin. Genel kural
  veya yakın konu bulunması, önceden seçilmiş sıkı ölçütü karşılamadı.
- **AYM:** Yeni adaylardan biri siyasi parti üyesinin hakaret mahkûmiyeti ve
  başvuru yollarını tüketmeme; diğeri özel şirket taşeron işçisinin şüphe feshi.
  Kamu görevlisinin disiplin yaptırımına ilişkin soruyla aynı rejimde değiller.
  İkinci kararın aleyhe olması değil, rejim farkı eleme nedenidir.
- **AİHM:** Yeni belgeler 115.368 ve 245.771; eski belgeler 315.629 ve 229.926
  karakter. Dördü de 60.000 sınırını aştı. Başlık veya konu benzerliğinden uygunluk
  sonucu çıkarılmadı. Burada artış/azalış ölçülemedi.

## Sınırlılıklar ve sonraki ölçüm

Bu bir **örnekler üzerinde, kör olmayan sorgu stratejisi pilotudur**. Sorguları
seçen ve metinleri değerlendiren aynı asistandır; uzman onaylı veya bağımsız
referans etiketler yoktur. Aynı örnekler üretim promptunda bulunur; görülmemiş
sorularda modelin otomatik sorgu üretme yeteneği ölçülmedi. İki uygun Danıştay
kararı benzer gerekçelidir; iki bağımsız hukuk alanında başarı sayılmaz.

Bağlı MCP'nin araç açıklamaları eski talimatları gösteriyordu; çalışan binary
sürümü ayrıca doğrulanmadı. İki yöntem aynı bağlantı/arama altyapısında çalıştı;
bu rapor 0.3.1'in istemcide yeniden başlatılmış olduğunu veya prompta uyumunu
kanıtlamaz. Değerlendirici yeni sorguları açıkça gönderdi.

Sonraki kontrollü çalışma için: görülmemiş sorular, uzman kontrolündeki referanslar
ve AİHM için baştan ilan edilmiş daha yüksek/eşit okuma bütçesi gerekir. Bu turda
başarısız örneklere göre sorgular sonradan değiştirilmedi; böyle bir ayarlama yeni
bir deney olarak kaydedilmelidir.

[Makine okunur kayıt](../evals/relevance/runs/2026-09-13-query-comparison.json)
her adayın kimliği, resmî bağlantısı, hash'i, okuma aralıkları, etiketi ve gerekçe
konumlarını içerir; gerçek karar gövdeleri veya yerel viewer token'ları içermez.
Çalışma zamanı, promptlar, kişisel ayarlar ve eski pilot değiştirilmedi.
Canlı koşu CI'a eklenmedi; commit/push/publish yapılmadı.
