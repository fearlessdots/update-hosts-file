pkgname=update-hosts-file
pkgver=0.1.0
pkgrel=1
pkgdesc="A program that automatically updates the /etc/hosts file"
arch=('any')
url="https://github.com/fearlessdots/update-hosts-file"
license=('GPL3')
depends=('glibc' 'gcc-libs')
makedepends=('go')
source=("${pkgname}-${pkgver}.tar.gz::${url}/archive/refs/tags/v${pkgver}.tar.gz")
sha256sums=()

# prepare() {
# 	 cd "$pkgname-$pkgver"
# 	 make deps
# }

build() {
	cd "$pkgname-$pkgver"
	make build
}

package() {
	cd "$pkgname-$pkgver"
	make DESTDIR=$pkgdir install
}
