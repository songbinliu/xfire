package example

import (
	"fmt"
	"testing"
)

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
