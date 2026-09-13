# Uygunluk pilotu — 12 Eylül 2026

**Sonuç: İlk iki adayla sınırlı dört sorguda, bu oturumda uygunluğu doğrulanarak
sunulabilecek karar çıkmadı. Bu sonuç modelin hukuki doğruluğunu ölçen bir oran değildir.**

## Gerçek kaynaklar: 4 soru, 8 aday

Soru ve filtreler aramalardan önce [planda](../evals/relevance/live-plan.json)
kaydedildi. Her kaynakta tek sayfa, `page_size=2`; otomatik sayfalama, sorgu
genişletme veya yeniden deneme yapılmadı. Aynı Go 0.3.0 MCP süreci ve ortak 3 saniye
sırası kullanıldı. Mevcut asistan değerlendirdi; başka AI API'si çağrılmadı.

| Kaynak | Tam okunan | Uygun | İlgisiz | Metinsel belirsizlik | Erişim hatası | Okuma tamamlanmadı |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Yargıtay | 2 | 0 | 0 | 2 | 0 | 0 |
| Danıştay | 1 | 0 | 1 | 0 | 1 | 0 |
| AYM | 1 | 0 | 1 | 0 | 1 | 0 |
| AİHM | 0 | 0 | 0 | 0 | 0 | 2 |
| Toplam | 4 | 0 | 2 | 2 | 2 | 2 |

Sekiz belge isteğinin altısı yanıt verdi; model dördünün tam metnini okudu.
Dolayısıyla erişim ve tam okuma aynı metrik değildir. Tam okunan dört adayda iki
ilgisiz ve iki belirsiz değerlendirmesi var. Kalan dört adaya hukuki etiket verilmedi.
Hiç karar sunulmadığından sunum precision'ı tanımsızdır; recall için bağımsız
ilgili-karar listesi yoktur. “0 uygun aday” gözlemi “%0 model doğruluğu” da değildir.

### Eleme gerekçeleri ve kaynaklar

- **Yargıtay / imzalı-ihtirazi kayıtsız bordro ve tanıkla ispat:**
  [2026/2266 E., 2026/4307 K.](https://karararama.yargitay.gov.tr/getDokuman?id=1224717400)
  ve [2026/2743 E., 2026/4308 K.](https://karararama.yargitay.gov.tr/getDokuman?id=1224719800)
  tam okundu. Fazla çalışma ve tanık/bordro kelimeleri bulunsa da dar sorunun ayırt
  edici bordro koşulları ve bunlara özgü ispat gerekçesi doğrulanmadı; belirsiz tutuldu.
- **Danıştay / sevkin tarh zamanaşımına etkisi:**
  [2025/21 E., 2026/1 K.](https://karararama.danistay.gov.tr/getDokuman?arananKelime=&id=1203795900)
  için bu turdaki istek zaman aşımına uğradı. Önceki oturumun bilgisi yeni tam
  okuma gibi sayılmadı. [2024/5940 E., 2025/5843 K.](https://karararama.danistay.gov.tr/getDokuman?arananKelime=&id=1218379600)
  tam okundu; uyuşmazlık avukatlar için günlük hasılat tespiti düzenlemesi hakkında,
  sevkin zamanaşımına etkisi hakkında değil. AND kelime eşleşmesi yeterli olmadı.
- **AYM / kamu görevlisinin sosyal medya paylaşımı için disiplin cezası:**
  [2014/5376 başvurusu](https://kararlarbilgibankasi.anayasa.gov.tr/kbb/pages/search/BireyselBasvuru?id=a2JiOjZiMmMyMTI0LWRlNWEtZmVhMC1kZjZlLWI4ODRiYTczMmVhZQ&type=BireyselBasvuru)
  tam okundu (§1, §26, §50 ve hüküm). Konu şeref/itibar hakkı ve erişim engeli
  kararlarının uygulanması; aranan disiplin-cezası sorusuyla aynı değil.
  [2023/71304 başvurusu](https://kararlarbilgibankasi.anayasa.gov.tr/kbb/pages/search/BireyselBasvuru?id=a2JiOjc3NDkzZGE1LWNjODctNDcyMC1iMWI2LWIzMzI2ZDNiOTkyZQ&type=BireyselBasvuru)
  belge isteği zaman aşımına uğradı.
- **AİHM / ilk sorguda avukat ve ifadenin mahkûmiyette kullanımı:**
  [001-252214](https://hudoc.echr.coe.int/eng?i=001-252214) 315.629,
  [001-251027](https://hudoc.echr.coe.int/eng?i=001-251027) 229.926 karakterdi.
  Boyut kontrolü için birer karakterlik parça alındı; tam metinler okunmadı.
  Bu nedenle uygun/ilgisiz sonucu yok. Başlıklardan çıkarım yapılmadı.

Hash'ler, belge kimlikleri, gerekçe konumları ve durumlar
[canlı sonuç kaydında](../evals/relevance/runs/2026-09-12-live.json) bulunur.
Gerçek karar gövdeleri dosyalara, uygulama loglarına veya fixture'lara kaydedilmedi.

## Sentetik örnekler: yalnız iç tutarlılık kontrolü

16 kurgusal örnekte aynı asistanın değerlendirmesi referans etiketleriyle **16/16**
uyumlu: 5 uygun, 7 ilgisiz, 4 belirsiz. Referansa göre uygunsuz sunum yok; eksik
metinler sunulmadı. Taraf iddiası, karşı oy, usulden ret, farklı hukuki rejim,
kaynak içi talimat ve “aleyhe sonuç ≠ ilgisiz karar” örnekleri bulunur.

**Örnek yazarı, referans etiketleyici ve değerlendirici aynı asistandır. Bu kör,
bağımsız veya uzman doğrulamalı bir model testi değildir. %100 hukuki başarı denemez.**
Sayısal çıktı yalnız sağlanan etiketlerle uyumu hesaplar; anlam doğruluğunu ölçmez.
Puanlayıcının altı birim testi geçti. API/model çağrısı yapmayan tekrar komutu
ve yeni istemciyle değerlendirme yöntemi [eval rehberindedir](../evals/relevance/README.md).

## Ne öğrendik, neyi henüz ölçmedik?

1. İlk iki arama adayını doğrudan emsal diye sunmak bu dar sorularda işe yaramadı.
   Geniş ifadelerle seçilen bu örnek, sorgu planlamasının ayrıca ölçülmesi gerektiğini
   gösterir; bütün arşiv veya adaptif model davranışı hakkında sonuç değildir.
2. Metni alma hatası, metindeki belirsizlik ve ilgisizlik ayrı durumlar olmalı.
   AİHM'de belge uzunluğu ayrı bir inceleme bütçesi gerektiriyor. Bu pilotta üst
   sınır önceden belirlenmemişti; uzun belgeleri okumama kararı yöntem sınırlılığıdır.
3. Modelin yanlış emsal sunma oranını henüz güvenilir biçimde ölçmedik. Bunun için
   uzman onaylı uygun/uygunsuz referanslar, bunları görmemiş bir değerlendirici,
   bağımsız gerekçe incelemesi ve tekrarlı çalıştırmalar gerekir.
4. Bir sonraki deneyde aynı sorular için sorgu stratejileri ve belge okuma bütçesi
   önceden sabitlenmeli; tüm koşullarda aynı aday/istek bütçesi kullanılmalı.

Bu turda Go çalışma zamanı, promptlar, kişisel MCP ayarları ve yayın akışı
değiştirilmedi. Test verileri/puanlayıcı/rapor eklendi; commit veya yayın yapılmadı.
