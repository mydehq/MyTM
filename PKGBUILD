# Maintainer: Soymadip <soumadip@zohomail.in>
pkgname=mytm
pkgdesc="Global theme manager plugin for MyCTL"
pkgver=0.0.0
pkgrel=1
arch=('x86_64')
url="https://github.com/mydehq/${pkgname}"
license=('GPL3')
sha256sums=('SKIP')

depends=(
  "myctl"
)

makedepends=(
  "go"
  "git"
)

source=("${pkgname}::git+file://${PWD}")

prepare() {
    ls -R "${srcdir}"
    cd "${srcdir}/${pkgname}/app"
    go mod tidy
}

build() {
    cd "${srcdir}/${pkgname}/app"
    go build -o "${pkgname}"
}

package() {
    install -Dm755 "${srcdir}/${pkgname}/app/${pkgname}" "${pkgdir}/usr/bin/${pkgname}"
}
