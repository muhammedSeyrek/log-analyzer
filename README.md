# Log Analyzer & Threat Detection Tool
Bu proje, sistem loglarını (Windows Event Logs, Linux Syslog/Journald ve macOS Logs) analiz ederek potansiyel güvenlik tehditlerini tespit eden, Go (Golang) dili ile geliştirilmiş çapraz platform (cross-platform) bir güvenlik aracıdır.

Proje, hem statik dosya analizi (geçmiş loglar) hem de canlı sistem izleme (real-time monitoring) yeteneklerine sahiptir.

## Teknik Özellikler
Dil: Go 

Mimari: Cross-Platform (Windows, Linux, macOS uyumlu)

Eşzamanlılık (Concurrency): I/O işlemlerinde ve hata yönetiminde Go Routines kullanılarak performans optimizasyonu sağlanmıştır. Standart çıktı (stdout) ve hata çıktıları (stderr) eşzamanlı işlenir.

## Log Kaynakları:

Windows: PowerShell üzerinden Security Event Log akışı.

Linux: journalctl (systemd) ve /var/log dosyaları (Syslog/Auth).

macOS: Apple Unified Log System (log stream).

## Derleme Yöntemi: Docker, bir "Build Environment" olarak kullanılarak hedef işletim sistemi için binary dosyalar üretilir. Hedef makinede Go kurulu olması gerekmez.

# Kurulum ve Çalıştırma
Bu proje, yerel makinenize Go kurulumu yapmanızı gerektirmez. Docker kullanılarak proje derlenir ve işletim sisteminize uygun, bağımsız çalışabilir (standalone) bir dosya üretilir.

Aşağıda işletim sisteminize uygun adımları takip ediniz.

1. Linux (Debian, Ubuntu, Kali, CentOS)
Linux sistemlerde program, doğrudan çekirdek (kernel) loglarına erişmek için journalctl veya /var/log kaynaklarını kullanır. Bu nedenle Root yetkisi gerektirir.

Bash
# 1. Docker kullanarak Linux uyumlu binary dosyasını oluşturun
docker build --build-arg TARGETOS=linux -t analyzer-builder .

# 2. Üretilen dosyayı almak için geçici bir konteyner oluşturun
docker create --name temp-container analyzer-builder

# 3. Dosyayı konteynerden dışarı aktarın
docker cp temp-container:/log-analyzer ./log-analyzer

# 4. Geçici konteyneri temizleyin
docker rm temp-container

# 5. Dosyaya çalıştırma izni verin
chmod +x log-analyzer

# 6. Uygulamayı ROOT yetkisiyle başlatın (Sistem loglarına erişim için zorunludur)
sudo ./log-analyzer
2. Windows (10/11/Server)
Windows üzerinde Güvenlik Loglarına (Security Event Log) erişim, Yönetici (Administrator) yetkileri gerektirir.

Bash
# 1. Windows uyumlu .exe dosyasını oluşturun
docker build --build-arg TARGETOS=windows --build-arg TARGETARCH=amd64 -t analyzer-win .

# 2. Geçici taşıyıcı oluşturun
docker create --name temp-win analyzer-win

# 3. Dosyayı .exe uzantısıyla dışarı aktarın
docker cp temp-win:/log-analyzer ./log-analyzer.exe

# 4. Temizlik
docker rm temp-win
Çalıştırma: Oluşturulan log-analyzer.exe dosyasına sağ tıklayın ve "Yönetici Olarak Çalıştır" seçeneğini kullanın.

3. macOS (Apple Silicon & Intel)
macOS üzerinde çalıştırırken işlemci mimarisine (M1/M2 vs Intel) dikkat edilmelidir.

Apple Silicon (M1, M2, M3 vb.) için:

Bash
docker build --build-arg TARGETOS=darwin --build-arg TARGETARCH=arm64 -t analyzer-mac-m1 .
docker create --name temp-mac analyzer-mac-m1
docker cp temp-mac:/log-analyzer ./log-analyzer-mac
docker rm temp-mac
chmod +x log-analyzer-mac
sudo ./log-analyzer-mac
Intel İşlemcili Mac'ler için:

Bash
docker build --build-arg TARGETOS=darwin --build-arg TARGETARCH=amd64 -t analyzer-mac-intel .
docker create --name temp-mac analyzer-mac-intel
docker cp temp-mac:/log-analyzer ./log-analyzer-mac
docker rm temp-mac
chmod +x log-analyzer-mac
sudo ./log-analyzer-mac
Yapılandırma
Tehdit algılama kuralları config/rules.yaml dosyasında tanımlanmıştır. Program yeniden derlenmeden bu kurallar değiştirilebilir.

Örnek Kural Yapısı:

YAML
rules:
  - name: "SSH Brute Force"
    pattern: "Failed password"
    type: "security"
  - name: "Windows Logon Failure"
    pattern: "failed to log on"
    type: "security"


# Raporlama ve Çıktı Yönetimi
Uygulama, statik analiz sonuçlarını kalıcı hale getirmek için CSV formatında raporlama desteği sunar.

Dizin Yapısı: Tüm rapor dosyaları, uygulamanın çalıştığı kök dizin altında otomatik olarak oluşturulan report/ klasörü içerisinde saklanır.

Dosya İsimlendirme: Raporların karışmasını önlemek ve tarihsel takibi sağlamak amacıyla dosya isimlerinde oluşturulma anına ait zaman damgası (timestamp) kullanılır.

Örnek Dosya Yolu: report/report_20260131_143000.csv