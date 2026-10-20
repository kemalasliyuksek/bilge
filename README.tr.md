# bilge

[![CI](https://github.com/kemalasliyuksek/bilge/actions/workflows/ci.yml/badge.svg)](https://github.com/kemalasliyuksek/bilge/actions/workflows/ci.yml)

[English](README.md) | Türkçe

Docker Engine'inizi izlemek, yönetmek ve temizlemek için sade bir masaüstü
uygulaması.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/screenshots/overview-dark.png">
  <img alt="bilge genel bakış ekranı: kalacak, incelenmesi gereken ve güvenle silinebilecek olarak ayrılmış disk kullanımı, engine bilgileri, Compose projeleri ve en büyük image ve volume'ler" src="docs/screenshots/overview-light.png">
</picture>

bilge, engine'inizde neler olduğunu ve diskte ne kadar yer kapladığını
gösterir, container'ları başlatıp durdurmanızı sağlar ve hâlâ ihtiyacınız olan
şeylere dokunmadan gerçekten artık olanları siler. Linux, macOS ve Windows'ta
çalışır; yerel engine'e ya da `DOCKER_HOST` ile belirtilene bağlanır.

> Durum: erken geliştirme aşamasında. Güvenmeden önce, bir şey kaybetmeyi göze
> alabileceğiniz bir engine üzerinde deneyin.

## Neler yapar

**İzleme.** Genel bakış ekranı Docker'ın disk kullanımını kalacak, incelenmesi
gereken ve silinebilecek olarak ayırır. Compose projelerinizi listeler; buna
container'ları silinmiş ama volume ve image'leri hâlâ duran projeler de dahildir.
En büyük image ve volume'leri de gösterir. Container, image, volume ve network'ler
için ayrı listeler vardır; her birinde arama, durum filtreleri ve sıralanabilir,
genişliği ayarlanabilir kolonlar bulunur.

**Yönetim.** Container'ları listeden başlatın, durdurun ve yeniden başlatın.

![Durum, yazılabilir katman boyutu ve başlat, durdur, yeniden başlat butonlarıyla container listesi](docs/screenshots/containers.png)

**Temizlik.** `docker system prune` ve çoğu "prune" butonu, o an hiçbir şeyin
kullanmadığı her şeyi silinebilir sayar. Hafta sonu durdurduğunuz bir Compose
yığınının veritabanı volume'ü de buna dahildir. bilge ise adayları iki gruba
ayırır:

- **Safe (güvenli)**: dangling image'ler, anonim volume'ler, Compose'a ait
  olmayan durmuş container'lar, boş network'ler ve kullanılmayan build cache.
  Tek bir onayla birlikte silinir. Build cache varsayılan olarak seçili gelmez,
  çünkü silinmesi sonraki build'lerin sıfırdan başlaması demektir.
- **Review (inceleme)**: isimli volume'ler, etiketli image'ler ve bir Compose
  projesine ait durmuş container'lar. Gerekçesiyle gösterilir ve tek tek
  silinir. İsimli bir volume, adı yazılarak onaylanır.

![Safe ve review gruplarıyla temizlik ekranı](docs/screenshots/cleanup.png)

Kullanımda olan ya da `bilge.keep=true` etiketi taşıyan hiçbir şey silinmek
üzere önerilmez. bilge silmeden hemen önce engine'i yeniden tarar ve siz
baktıktan sonra değişen her kalemi atlar; örneğin bir container'ın az önce
bağladığı bir volume'ü. Kuralların tamamı (İngilizce)
[docs/design.md](docs/design.md) dosyasındadır.

![İsimli bir volume için adının yazılmasını isteyen onay penceresi](docs/screenshots/remove-volume.png)

## İndirme

Linux, macOS (universal) ve Windows derlemeleri her
[release](https://github.com/kemalasliyuksek/bilge/releases) sayfasında bir
`SHA256SUMS` dosyasıyla birlikte yer alır. Dosyalar henüz imzalı değildir:
macOS ilk açılışta onay ister (sağ tık, Aç), Windows SmartScreen de uyarı
gösterebilir. Linux'ta bilge WebKitGTK 4.1 gerektirir (Debian ve Ubuntu'da
`libwebkit2gtk-4.1-0`).

## Geliştirme

Gerekenler:

- Go 1.25+
- Node.js 24 ve pnpm
- [Wails v2](https://wails.io) CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- [just](https://just.systems) ve [golangci-lint](https://golangci-lint.run) v2
- Linux: `libgtk-3-dev` ve `libwebkit2gtk-4.1-dev`

[mise](https://mise.jdx.dev) kullanıyorsanız `mise install`, `mise.toml`
dosyasındaki sabitlenmiş araç sürümlerini kurar.

```sh
cd frontend && pnpm install && cd ..
just dev      # canlı yenilemeyle çalıştır
just check    # format, lint, tip denetimi, testler
just build    # build/bin altına production derlemesi
```

Kod, commit mesajları ve dokümanlar İngilizcedir; katkılar da İngilizce
beklenir.

## Nasıl geliştiriliyor

bilge'i Kemal Aslıyüksek tasarlıyor ve sürdürüyor. Kodun çoğu, bir AI kodlama
ajanı olan Claude Code ile [docs/design.md](docs/design.md) dosyasındaki
tasarıma göre yazılıyor. Yönü Kemal belirliyor, kararları o veriyor ve her
değişikliği o inceliyor. Değişiklikler kalite kapısından geçmek zorunda,
davranış değişikliklerine test ekleniyor.

## Lisans

[MIT](LICENSE)

Docker, Docker, Inc.'in ticari markasıdır. bilge bağımsız bir projedir;
Docker, Inc. ile bağlantılı değildir ve onun tarafından desteklenmemektedir.
