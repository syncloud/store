package machine

import (
	"log"
	"runtime"
	"syscall"
)

var DPKGArch = dpkgArchFromGoArch(runtime.GOARCH)

func dpkgArchFromGoArch(goarch string) string {
	goArchMapping := map[string]string{
		// go      dpkg
		"386":     "i386",
		"amd64":   "amd64",
		"arm":     "armhf",
		"arm64":   "arm64",
		"ppc":     "powerpc",
		"ppc64":   "ppc64", // available in debian and other distros
		"ppc64le": "ppc64el",
		"riscv64": "riscv64",
		"s390x":   "s390x",
	}

	if goarch == "arm" {
		if MachineName() == "armv6l" {
			return "armel"
		}
	}

	dpkgArch := goArchMapping[goarch]
	if dpkgArch == "" {
		log.Panicf("unknown goarch %q", goarch)
	}

	return dpkgArch
}

func MachineName() string {
	var u syscall.Utsname
	err := syscall.Uname(&u)
	if err != nil {
		return "unknown"
	}

	buf := make([]byte, len(u.Machine))
	for i, c := range u.Machine {
		if c == 0 {
			buf = buf[:i]
			break
		}
		// c can be uint8 or int8 depending on arch (see comment above)
		buf[i] = byte(c)
	}

	return string(buf)
}
