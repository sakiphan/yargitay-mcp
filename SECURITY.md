# Güvenlik ve veri işleme

Bu uygulama bağımsız, yerel karar erişim yazılımıdır; Yargıtay veya model
sağlayıcılarıyla resmî bağlantısı yoktur. Varsayılan mimaride proje sahibine sorgu,
karar veya telemetri gönderilmez. Kullanıcı istekleri doğrudan resmî siteye gider.
Karar metni bellekte tutulur; model istemcisinin gönderim/saklama politikası ayrıdır.

Kamuya açık veriler yine kişisel veri olabilir. Kodun açık kaynak olması veya AGPL
lisansı veri işleme, yurt dışına aktarım ve otomatik erişim yükümlülüklerini ortadan
kaldırmaz. Kaynak metin uydurulmamalı; kişisel veri içeren örnekler issue/fixture'a
konmamalıdır. HTML temizleme tek başına prompt injection koruması değildir.

Kaynaklar: [FSEK m.31](https://agri.ktb.gov.tr/Eklenti/135532%2C5846-sayili-fikir-ve-sanat-eserleri-kanunupdf.pdf?0=),
[KVKK alenileştirme açıklaması](https://www.kvkk.gov.tr/Icerik/6843/-ALENILESTIRME-HAKKINDA-KAMUOYU-DUYURUSU),
[KVKK yurt dışına aktarım rehberi](https://www.kvkk.gov.tr/Icerik/8143/Kisisel-Verilerin-Yurt-Disina-Aktarilmasi-Rehberi).
Somut kullanım ve lisans uyuşmazlıkları için hukuki değerlendirme ayrıca gerekir.

Yayın öncesi maintainer GitHub private vulnerability reporting'i etkinleştirmelidir.
Etkinliği bu yerel geliştirmede doğrulanmadı. Güvenlik açığını gerçek karar metniyle
herkese açık issue'ya koymayın. Aktif özel kanal yoksa maintainer'dan özel iletişim
kanalı isteyin; hassas ayrıntıları kamuya göndermeyin.
