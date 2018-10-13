# sdktest
golang http api sdk test tool
> a train of thought

# Demo
``` go
import (
	"testing"
)

func Test_Demo(t *testing.T) {
	st := NewSDKTester(t, "xml")
	respData := []byte(`<xml><code>ok</code></xml>`)

	reqWant := map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	}

	st.HandleHTTP("/uri", respData, reqWant)
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

	respWant := map[string]interface{}{
		"Code": "ok", // Field name of struct filed, not tag name
	}

	st.CheckResponse(response, respWant)
}
```
