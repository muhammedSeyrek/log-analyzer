# Log Analyzer & Threat Detection Framework


### Projenin özetini geçmeden, nasıl çalıştırılır?
  •	make docker-run ile çalıştırılabilir,
  •	make extract docker ile izole linux ortamı,
  •	make demo ve make advanced-demo Analiz yapılabilir, 
  •	make live ile canlı Analiz yapılabilir (Windows ise bazı kısıtlamaalar kaldırılmalı, saldırı yapılacaksa firewall kapatılabilir.)
  •	make clean ile herşey eski haline döndürülebilir.

Bu proje, sistem loglarını (Windows, Linux, macOS) analiz ederek güvenlik tehditlerini tespit etmek amacıyla geliştirilmiş, Go tabanlı ve eklenti (plugin) mimarisine sahip bir güvenlik aracıdır. Proje, hem geçmişe dönük statik analiz hem de gerçek zamanlı (live) izleme yetenekleri sunar.

# Teknik Özellikler
Plugin Tabanlı Mimari: Yeni analiz modülleri plugins/ dizinine eklenerek CLI'a otomatik entegre edilebilir.

Çapraz Platform: Docker tabanlı derleme sistemi ile Windows, Linux ve macOS için bağımsız çalıştırılabilir (binary) dosyalar üretilir.

Tehdit Algılama: config/rules.yaml üzerinden özelleştirilebilir pattern eşleşmeleri.

Eşzamanlı İşleme: Go Routines ile yüksek performanslı log tarama ve I/O yönetimi.

Raporlama: Tespit edilen tehditlerin zaman damgalı CSV formatında dışa aktarımı.

# Kurulum ve Derleme
Docker ile Derleme
Yerel bir bağımlılık kurmadan, hedef işletim sistemi için binary dosyalarınızı Docker üzerinden üretebilirsiniz.

Bash
# Linux binary'sini derleyip ./out dizinine çıkartmak için:
make extract
Makefile ile Yerel Geliştirme
Geliştirme ortamınızda Go kuruluysa, Makefile üzerindeki otomasyonları kullanabilirsiniz:

Bash
make build   # Plugin listesini günceller ve projeyi derler
make list    # Kayıtlı plugin'leri listeler
make clean   # Derleme artıklarını temizler
Kullanım
Uygulama, plugins/ dizinindeki modülleri otomatik olarak algılar. Ana eklentimiz olan loganalyzer üç farklı modda çalışır:

Statik Analiz
Diskteki mevcut log dosyalarını tarar ve bulguları raporlar.

Bash
# Hızlı demo (dummy.log üzerinden)
make demo

# Özel bir dosya tarama ve rapor oluşturma
./log-analyzer loganalyzer static --file advanced_dummy.log --report
[Ekran Görüntüsü Yeri: Terminal Statik Analiz Sonuçları]
Açıklama: Terminal üzerinde yakalanan SSH ve Sudo tehditlerinin görünümü.

Canlı İzleme (Live Mode)
Sistemi anlık olarak dinler ve bir tehdit tespit edildiğinde uyarı verir. (Windows Security Log veya Linux /var/log erişimi için Administrator/Root yetkisi gerektirir).

Bash
make live
[Ekran Görüntüsü Yeri: Canlı İzleme Ekranı]
Açıklama: Sistem loglarının anlık takibi sırasında yakalanan alarmlar.

# Raporlama
Başarılı bir statik analizden sonra --report parametresi kullanıldığında, bulgular reports/ dizini altında saklanır.

Dosya Formatı: report_YYYY-MM-DD_HH-MM-SS.csv

İçerik: Kural Adı, Tehdit Tipi, Eşleşen Log Satırı ve Zaman Damgası.

[Ekran Görüntüsü Yeri: CSV Rapor Örneği]
Açıklama: Oluşturulan bir CSV raporunun Excel veya metin düzenleyicideki görünümü.

# Test Senaryoları
Projeyle birlikte gelen advanced_dummy.log dosyasını taratarak aşağıdaki tespit kurallarını simüle edebilirsiniz:

Brute Force: Çoklu başarısız SSH giriş denemeleri.

Privilege Escalation: Şüpheli sudo yetki kullanımları.

Persistence: Yeni kullanıcı ekleme ve güvenlik gruplarına atama işlemleri.
# Stream Processor (Gerçek Zamanlı Veri İşleme Modülü)

streamprocessor, ana iskelete plugin mimarisi üzerinden eklenmiş ikinci bir eklentidir. Amacı, yüksek hacimli bir log/trafik akışını diske veya belleğe tamamını yüklemeden, satır satır gelirken anlık olarak parse edip filtrelemektir. Veriyi standart girdiden (stdin) okur; bu sayede herhangi bir kaynaktan boru (pipe) ile beslenebilir ve Docker konteyneri içinde ek bir bağlama (mount) gerektirmez.

Modül, loganalyzer ile aynı plugin sözleşmesini (Name / Description / Command) uygular ve go generate adımıyla CLI'a otomatik kaydolur; ana koda hiçbir müdahale yapılmadan sisteme dahil olur.

## netmetrics eklentisi (Scapy + UDP)

Uc uncu eklenti `netmetrics`, Scapy ile ag paketlerini yakalar ve metrikleri
UDP uzerinden uzak bir sunucuya push eder; sunucu tarafi istenirse Slack'e
iletir. Ayrintili kullanim icin: NETMETRICS.md

    docker compose up --build netmetrics-collector netmetrics-agent
    log-analyzer list   # -> netmetrics gorunur

## Çalışma Mantığı

Modül bir üretici-tüketici hattı (producer-consumer pipeline) olarak kurgulanmıştır:

Okuyucu, girdiyi bufio.Scanner ile tek seferde bir satır okur ve sınırlı kapasiteli (bounded) bir kanala yazar. Bir veya daha fazla işçi (worker) goroutine bu kanaldan satırları alır, config/rules.yaml içindeki kurallara göre eşleştirir ve eşleşen satırı anında çıktıya yazıp bırakır.

Bellek kullanımının sabit kalmasının üç nedeni vardır: akışın tamamı hiçbir zaman bellekte tutulmaz, yalnızca işlenmekte olan satırlar bellektedir; kanal sınırlı kapasiteli olduğu için işçiler yavaş kaldığında okuyucu otomatik olarak bekler (backpressure) ve kuyruk sınırsız büyüyemez; eşleşen satırlar biriktirilmez, anında yazılıp atılır. Bu nedenle akış ister bir milyon ister sonsuz satır olsun, bellek tüketimi aynı kalır.

İşçiler eşzamanlı çalıştığından, sayımlar her işçinin kendi yerel sayaçlarında tutulur ve akış bittiğinde tek seferde birleştirilir. Böylece paylaşılan veriye eşzamanlı yazımdan kaynaklanan kilit maliyeti ve yarış durumu (race condition) ortadan kalkar.

## Kullanım

Varsayılan davranış, düşük bellek tüketen stream modudur. Sadece logları incelemek isteyen bir kullanıcı herhangi bir bayrak vermeden çalıştırdığında otomatik olarak bu modda, sade çıktıyla çalışır: yalnızca eşleşen satırlar ve kısa bir özet gösterilir.

```bash
# Bir dosyayı işleme
cat dummy.log | log-analyzer streamprocessor run

# Sentetik trafik üretip aynı anda işleme
log-analyzer streamprocessor gen --count 1000000 | log-analyzer streamprocessor run
```

Detaylı tanılama çıktısı (bellek kullanımı, saniyedeki satır sayısı, kural ve tip bazında eşleşme dökümü) yalnızca --verbose bayrağıyla istenir:

```bash
log-analyzer streamprocessor gen --count 1000000 | log-analyzer streamprocessor run --verbose
```

## Sentetik Trafik Üreteci

gen alt komutu, ayrı bir test aracına ihtiyaç duymadan, kurallara uyan ve uymayan satırların karışımından oluşan gerçekçi sahte trafik üretir. Üreteç de tek seferde bir satır ürettiği için kendisi de bellekte büyümez.

```bash
log-analyzer streamprocessor gen --count 1000000        # 1 milyon satır üret
log-analyzer streamprocessor gen --count 0               # sonsuz akış (Ctrl+C ile durur)
log-analyzer streamprocessor gen --rate 5                # saniyede 5 satır (canlı akış hissi)
log-analyzer streamprocessor gen --match-ratio 0.3       # satırların %30'u bir kurala uysun
```

## Bellek Karşılaştırması (stream ve buffered)

Modülde, tasarımın faydasını ölçülebilir biçimde göstermek için iki mod bulunur. Varsayılan stream modu sabit bellekte çalışır. buffered modu ise eski yaklaşımı temsil eder: eşleşen tüm satırları bellekte biriktirip işlem sonunda yazar, dolayısıyla bellek kullanımı eşleşme sayısıyla doğru orantılı büyür. İki mod da aynı girdide aynı sonucu üretir; tek fark bellek davranışıdır.

```bash
# Sabit bellek
log-analyzer streamprocessor gen --count 5000000 --match-ratio 1.0 | log-analyzer streamprocessor run --verbose

# Bellek akışla birlikte büyür
log-analyzer streamprocessor gen --count 5000000 --match-ratio 1.0 | log-analyzer streamprocessor run --verbose --mode buffered
```

Aynı beş milyon eşleşen satırda stream modu birkaç megabaytta sabit kalırken, buffered modunun bellek tüketimi yüzlerce megabayta çıkar. Akış büyüdükçe bu fark doğrusal olarak açılır.

## Bayraklar

```
--mode, -m         İşleme modu: stream (varsayılan) veya buffered
--workers, -w      İşçi goroutine sayısı (stream modu, varsayılan 1)
--buffer, -b       Kanal kapasitesi (stream modu, varsayılan 1000)
--filter, -f       Kurallara ek olarak aranacak metin (opsiyonel)
--quiet, -q        Eşleşen satırları gizle, sadece özet göster
--verbose, -V      Tam tanılama raporu (bellek, hız, kural/tip dökümü)
--config, -c       Kural dosyası yolu (varsayılan config/rules.yaml)
```

## Docker

streamprocessor, Dockerfile'da ayrı bir işlem gerektirmez; mevcut go generate ve derleme adımları sayesinde imaja otomatik dahil olur. Konteyner içinde veri akışını boru ile kurmak için entry point'in kabuk (sh) ile çağrılması yeterlidir:

```bash
make docker-stream

# veya doğrudan:
docker run --rm --entrypoint sh log-analyzer:latest -c \
  "log-analyzer streamprocessor gen --count 1000000 | log-analyzer streamprocessor run -q -V --workers 4"
```