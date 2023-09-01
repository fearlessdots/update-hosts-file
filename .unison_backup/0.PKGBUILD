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
sha256sums=('a4adff16ecd2fc8dcb5cd0bbb467f1ed92151963179e4fa922911c9588542de9')

build() {
	cd "$pkgname-$pkgver"
	make build
}

package() {
	# Create directories for shell autocompletion
	echo "Creating directories for shell autocompletion"
	mkdir -p ${pkgdir}/usr/share/bash-completion/completions ${pkgdir}/usr/share/zsh/site-functions \
		${pkgdir}/usr/share/fish/vendor_completions.d

	cd "$pkgname-$pkgver"
	make DESTDIR=$pkgdir install
}
