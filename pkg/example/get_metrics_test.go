package example

import (
	"fmt"
	"testing"
)

func Test_ConvertPodUID(t *testing.T) {
	inputs := []string{
		"kubernetes://video-671194421-vpxkh.default",
		"kubernetes://inception-be-41ldc.default",
		"kubernetes://inception-be-41ldc.istio-system",
		"kubernetes://inception-be-41ldc.istio-system.",
	}

	expects := []string{
		"default/video-671194421-vpxkh",
		"default/inception-be-41ldc",
		"istio-system/inception-be-41ldc",
		"istio-system/inception-be-41ldc",
	}

	for i := range inputs {
		ain := inputs[i]
		aout, err := convertPodUID(ain)
		if err != nil {
			t.Errorf("convert UID: %v failed: %v", ain, err)
		}

		if aout != expects[i] {
			t.Errorf("Not equal: %v Vs. %v", aout, expects[i])
		}
	}
}

func Test_ConvertPodUID_Fail(t *testing.T) {
	inputs := []string{
		"netes://video-671194421-vpxkh.default",
		"//inception-be-41ldc.default",
		"kubernetes://inception-be-41ldc",
		"kubernetes://inception-be-41ldc. ",
		"kubernetes://inception-be-41ldc-istio-system",
		"kubernetes:// .inception-be-41ldc-istio-system",
	}

	for i := range inputs {
		_, err := convertPodUID(inputs[i])
		if err == nil {
			t.Errorf("convert UID should have failed with input: %v", inputs[i])
			return
		}

		fmt.Printf("input: %v, err: %v\n", inputs[i], err)
	}
}

func Test_ConvertSVCUID(t *testing.T) {
	inputs := []string{
		"productpage.default.svc.cluster.local",
		"productpage.istio.svc.cluster.local",
		"a.b.svc.cluster.local",
		"a.bb.svc.x.y",
		"aa.bb.svc",
	}

	expects := []string{
		"default/productpage",
		"istio/productpage",
		"b/a",
		"bb/a",
		"bb/aa",
	}

	for i := range inputs {
		ain := inputs[i]
		aout, err := convertSVCUID(ain)
		if err != nil {
			t.Errorf("convert UID: %v failed: %v", ain, err)
		}

		if aout != expects[i] {
			t.Errorf("Not equal: %v Vs. %v", aout, expects[i])
		}
	}
}

func Test_ConvertSVCUID_Fail(t *testing.T) {
	inputs := []string{
		"productpage.default.cluster.local",
		"productpage.istio.cluster.local",
		"a.svc.cluster.local",
		"a.bb..svc.x.y",
	}

	for i := range inputs {
		_, err := convertSVCUID(inputs[i])
		if err == nil {
			t.Errorf("convert UID should have failed with input: %v", inputs[i])
			return
		}

		fmt.Printf("input: %v, err: %v\n", inputs[i], err)
	}
}
