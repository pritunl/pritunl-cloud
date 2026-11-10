package compositor

import (
	"github.com/pritunl/pritunl-cloud/utils"
)

func GetEnv(username, driPath string, driPrime bool) (
	envData string, err error) {

	dsp, err := GetDisplay(username)
	if err != nil {
		return
	}

	env := []string{}
	if dsp.Display != "" {
		env = append(env, utils.FilterSystemdEnv("DISPLAY", dsp.Display))
	}
	if dsp.Xauthority != "" {
		env = append(env,
			utils.FilterSystemdEnv("XAUTHORITY", dsp.Xauthority))
	}
	if dsp.WaylandDisplay != "" {
		env = append(env,
			utils.FilterSystemdEnv("WAYLAND_DISPLAY", dsp.WaylandDisplay),
			utils.FilterSystemdEnv("SDL_VIDEODRIVER", "wayland"),
			utils.FilterSystemdEnv("GDK_BACKEND", "wayland"),
			utils.FilterSystemdEnv("XDG_SESSION_TYPE", "wayland"),
			utils.FilterSystemdEnv("XDG_SESSION_CLASS", "user"),
		)
		if dsp.XdgRuntime != "" {
			env = append(env,
				utils.FilterSystemdEnv("XDG_RUNTIME_DIR", dsp.XdgRuntime))
		}
	}
	env = append(env, utils.FilterSystemdEnv("SDL_GRAB_KEYBOARD", "0"))
	if driPath != "" {
		env = append(env,
			utils.FilterSystemdEnv("DRI_RENDER_DEVICE", driPath))
	}
	if driPrime {
		env = append(env,
			utils.FilterSystemdEnv("DRI_PRIME", "1"),
			utils.FilterSystemdEnv("__NV_PRIME_RENDER_OFFLOAD", "1"),
			utils.FilterSystemdEnv("__GLX_VENDOR_LIBRARY_NAME", "nvidia"),
		)
	}

	for _, line := range env {
		if line != "" {
			envData += "\n" + line
		}
	}

	return
}
