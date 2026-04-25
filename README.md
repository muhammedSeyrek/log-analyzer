# Log Analyzer & Threat Detection Framework
### Ayrıntılı Test için Report.pdf dosyasına bakılması rica olunur.

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