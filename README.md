# sdktest
golang http api sdk test tool
> a train of thought

# Demo
``` go
import (
	"testing"
)

func Test_Demo(t *testing.T) {
	respData := []byte(`<xml><code>ok</code></xml>`)

	reqWant := map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	}

	respWant := map[string]interface{}{
		"Code": "ok", // Field name of struct filed, not tag name
	}

    st := sdktest.NewSDKTester(t, sdktest.Options{
         RespType: "xml",
         XMLRoot:  "xml",
         RespData: respData,
         RespWant: respWant,
         ReqWant:  reqWant,
         URI:      "/pay/prepay",
     })

	client := NewClient(Option{
		BaseURL:        st.URL(),
		SkipVerifySign: true,
		//...
	})
    client.skipVerifySignature = false

	response, err := client.APIFunc1(APIFunc1Request{
		Field1: "value1",
		Field2: "value2",
	})

	if err != nil {
		t.Error(err)
		return
	}

	st.Test(response)
}
```
