package upload_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/mydms/features/upload"
)

func TestClientGet(t *testing.T) {
	path := "/upload/ID"
	var record MockAssertion
	mockServer := NewMockServer(&record, MockServerProcedure{
		URI:        path,
		HTTPMethod: http.MethodGet,
		Response: MockResponse{
			StatusCode: 200,
			Body: []byte(`{ "id": "f84b3381-6700-40ec-9f1c-084f42cb8daa",
			"fileName": "documentPlain.pdf",
			"payload": "JVBERi0xLjcKJeLjz9MKMSAwIG9iago8PC9PdXRsaW5lcyAyIDAgUi9QYWdlcyAzIDAgUi9UeXBlL0NhdGFsb2c+PgplbmRvYmoKNCAwIG9iago8PC9Db250ZW50cyA1IDAgUi9NZWRpYUJveFswIDAgNjEyLjAwMDAwMDAwMDAwMCA3OTIuMDAwMDAwMDAwMDAwXS9QYXJlbnQgMyAwIFIvUmVzb3VyY2VzPDwvRm9udDw8L0YxIDkgMCBSPj4vUHJvY1NldCA4IDAgUj4+L1R5cGUvUGFnZT4+CmVuZG9iago1IDAgb2JqCjw8L0xlbmd0aCAxMDc0Pj4Kc3RyZWFtCpqnrSnYU2NfK6P8HJu2KvQGEcCUcvelz77Y3SmhPeONySA8nwnG7dOvtf78F9dLXU8gckDVW+Omx4eAM5+q12u3GFP2JjQ1fJr4vg8GHN91jLZuiW45ukYQyYJqd4Lyln+TdEp8f8wbItNQOpgC7lbi+KHUM4gOpinafYETmCckG210nbMmgh2CRXGLOuiAbsoSHNEBYF68B02yws9BastIc0XR8TGzLIDUeiNfZAsTOz43NfshllJ6Xg/pxnZcLReNB1JzRuzATKUESKLMlOmkgEdN66owDoPdn3f9xKuHeOG+la3VwBRFOsDX4nXrxojAl0Ko8VJPywJT9S6AwNUrZmtzYHOvDqrPQjAqY/67qzNmspHPetVoseRLZ0Pw8ZzAbvV/VuLzc0ZIjo1EkUAsHRc5a/dPEqW2UXfw826u8sawqD1tbgARERtpiy1vYdLgRmRtZOCWMsRUr/MubM4J+QO1++ZMgTokEwdWIQUohPI4a6zaKi7tFsqkfm+6BdZnmiDg4+ubLZlyK903yUz/lNRofbgAMyNA+8XdRuqSxNGkPv5IdRiBSLf3ch/wFL/YNYEJO2qifrwVLI3GogJsfolPWHOC6nGJUgJHpOLeTL2Xf3F/+d2gAvOuNSOtPTJjCorCkAKWCfymfeIaICIOb0vWXHmOcsWqKL5zubifAxXcwTRVnV0WkQHETJgWFLwQOhmec2mT4hkSaeOZGoWiMErSImAY2d2P9AeqaCW0yWhRrPGsGTjh7f36R+8Cvlh9SeepaWmJb+JRH6cyxcZFYpv4Tx5N9BW1nNY0XI06k6OI8VFGukPEKvYu3H26k3m9B+ZzEIpIrSmmgjOhQPuhf9La/hstvA7TjZ90cLpl9KpRQhC/yFvi7sFApysNPjDc8qA8M3IdQAhgqkfxy+YdIhfGCCJjB3cHOcDBu2R6KNF5t+ggKdRCR1OzwjtvS7XMl9xxzDQ97tnV8n/ZEaylI4L54eyOOT35kQ82YKNGJ0EvzGv9lKNKRqPQQPsQz5cmaMKWGgAWOytGQc69wdwswiZNaBEM1LMpmXWs0uffZksddtvEF8+IE78raTIQVcmxh5bTVeMSDyDgjOHFYIQuVPRJZmgIryuAGMXRWhrgkDYexhESXKVa90C8so/NSHr1At1n811COb60jTsL5VVrWcouis/7zALJZCt/K2ebIDXxA2YrldxPSU4fO4sPrix/MqIjwV8FqXttcnBY1avadUIFkvj6aqs5NowU2DLI7stGCr5H/peyaYwSjqDH13RLTQL1TGmLOw8S6evN5HicIeyT5nU6kKKXR0OJnqe9HeVRU7FGUJzWKjhTAKMBgIAMolBdEZdGdXOrVqblcW+YgwDdDbTyJIEouXFzUREHDWkNt61LAXPRZwvVvY5aZz79Sv8BjX7pxws0KhAuCu7Hh2VuZHN0cmVhbQplbmRvYmoKOSAwIG9iago8PC9CYXNlRm9udC9IZWx2ZXRpY2EvRW5jb2RpbmcvV2luQW5zaUVuY29kaW5nL05hbWUvRjEvU3VidHlwZS9UeXBlMS9UeXBlL0ZvbnQ+PgplbmRvYmoKOCAwIG9iagpbL1BERi9UZXh0XQplbmRvYmoKNiAwIG9iago8PC9Db250ZW50cyA3IDAgUi9NZWRpYUJveFswIDAgNjEyLjAwMDAwMDAwMDAwMCA3OTIuMDAwMDAwMDAwMDAwXS9QYXJlbnQgMyAwIFIvUmVzb3VyY2VzPDwvRm9udDw8L0YxIDkgMCBSPj4vUHJvY1NldCA4IDAgUj4+L1R5cGUvUGFnZT4+CmVuZG9iago3IDAgb2JqCjw8L0xlbmd0aCA2NzY+PgpzdHJlYW0KtCg0OdebiiOslTrvTyL7V7FbwLnfXBVZ1bmkZzFO6PfKkHdkOq1fGAbow6Y3KNTg70uuMfHZoTFHiIZvuazgoqamOTX7+8PW2xn2QNV6ncSgSBnsYv2FhqIhXJSsV3Eag9CM0amoZg2QsOOxs7P2CGh7ZH96prPRDeo9Jak1AyxZ/xN6YxPkei8E9PFR8CpICqYVh+zZO7Z4m+iIOANmCdiXf6Ly7DwAdkzh6marI+zLgzZzrbExJ8etQci/f2PxmGAGf4wuAI8Sp7CG4yGVkUGohVijqUmkabtXmRZpno+RaWs6WjnOSidW9BcH3kOIGbph7WWKz9h3LI6P2FyyUcpORfjB2gbuqSY5Rr0g17zA/dKms2FqGGqp3MeRW9P3IeSiEX632Zyt6GgqprVO0PMGyB32smB2WUpnkrqjaNLKrMO4VP4IUkR2bSELZH8BUrRKXuhpyACgABRN7NTzytPhH4SdE9C6fEkuiNjFcHRXj1jyN+jsaqNUGLb7I8dZ6dIN1FShh2gfoOLr2h3VrOcoNCC5Qemc3xdoYc9s9lVakjRsaaaIJKlrbokrtxw04TANI+/IpDVwOFMQu32asMTtvHoi2ZZa8XyHWXwmJeOI9EhfeNxJz00D4Emvlyt6DgUwydifvclxneAYHN7DWKDL4J16iKixsdENQUejwsSRi6XMUcbgIGxI4E51gYMJx7uwaKjvz/3tLSIZGrYvqYYiia4Wcince/5pkHNJVW8mtd3y2tpCjgoZupvSbesEnnrh1zrsRQ8GBAc+N5ZEHhlnYgKLqkrv3b3k4ihd8NOpP0sONXjC2SC3Q5TF+9S5It6PBZkUILSh9aX8a8NkwFzNQY1eniDMxhLLAChdZOETmXvC8SbzEcmOb2W3lOlBkfHRxGVuZHN0cmVhbQplbmRvYmoKMyAwIG9iago8PC9Db3VudCAyL0tpZHNbNCAwIFIgNiAwIFJdL1R5cGUvUGFnZXM+PgplbmRvYmoKMiAwIG9iago8PC9Db3VudCAwL1R5cGUvT3V0bGluZXM+PgplbmRvYmoKMTAgMCBvYmoKPDwvQ3JlYXRpb25EYXRlPDI1N2M2YjNmYjIyZDU4YzQzNmMzNjk3ZTA0ODQyOGY0NmMxMjlkOWMwNjFhY2I+L0NyZWF0b3I8MzMyNzJmNmFhMDQxNDA5NDczODMyODc1MWU5OTZjYjIzMDBjYzNkZTQwNTg4MzY4ODUzMzJiMmUzNzM3ODRhMGZhMjVkMjcyPi9Nb2REYXRlPDI1N2M2YjNmYjIyZDU4YzQzNmMzNjk3ZTA0ODQyOGY0NmMxMjlkOWMwNjFhY2I+L1Byb2R1Y2VyPDExMjIzZjZjZjA2ODQ4OGEzN2Q5NmI2MTA1OTY3ZmEwMzE+Pj4KZW5kb2JqCjExIDAgb2JqCjw8L0NGPDwvU3RkQ0Y8PC9BdXRoRXZlbnQvRG9jT3Blbi9DRk0vVjIvTGVuZ3RoIDU+Pj4+L0ZpbHRlci9TdGFuZGFyZC9PPDYwNzY2NTExOTdkNGU4NWM5YmEyNjk4NzMzNjcxYTYyMWY1MmY4YzgzZGVmY2FlNjBiZTkxZWFkN2EzOTY4Yzg+L1AgLTM5MDEvUiAyL1N0bUYvU3RkQ0YvU3RyRi9TdGRDRi9VPDhiYjQzYjU5ZjRhNTliMmY4ZDEwZGE5NmZmMmU2MWIwNTA0NmY4ZjU1OTI3ZmIwZDY0N2UyYTBjY2M1ZDdhNWE+L1YgMT4+CmVuZG9iagp4cmVmCjAgMTIKMDAwMDAwMDAwMCA2NTUzNSBmIAowMDAwMDAwMDE1IDAwMDAwIG4gCjAwMDAwMDI0MDYgMDAwMDAgbiAKMDAwMDAwMjM0OSAwMDAwMCBuIAowMDAwMDAwMDc1IDAwMDAwIG4gCjAwMDAwMDAyMjcgMDAwMDAgbiAKMDAwMDAwMTQ3MyAwMDAwMCBuIAowMDAwMDAxNjI1IDAwMDAwIG4gCjAwMDAwMDE0NDYgMDAwMDAgbiAKMDAwMDAwMTM1MCAwMDAwMCBuIAowMDAwMDAyNDQ4IDAwMDAwIG4gCjAwMDAwMDI3MTMgMDAwMDAgbiAKdHJhaWxlcgo8PC9FbmNyeXB0IDExIDAgUi9JRFs8ZDJhOTI5YzBmMjI4YmI3YTY2MjVkZTgxNGVmMWM3Yzg+IDxkMmE5MjljMGYyMjhiYjdhNjYyNWRlODE0ZWYxYzdjOD5dL0luZm8gMTAgMCBSL1Jvb3QgMSAwIFIvU2l6ZSAxMj4+CnN0YXJ0eHJlZgoyOTc1CiUlRU9G",
			"mimeType": "application/pdf",
			"created": "2020-08-14T11:52:31.956394Z" }`),
		},
	})

	url := mockServer.URL
	client := upload.NewClient(url)
	u, err := client.Get("ID", "TOKEN")
	if err != nil {
		t.Errorf("cannot get ID '%s', %v", "ID", err)
	}

	assert.Equal(t, 1, record.Hits(path, http.MethodGet))
	// check request headers
	expHeaders := []http.Header{{
		"Authorization": []string{"Bearer TOKEN"},
		"Cache-Control": []string{"no-cache"},
	}}
	assert.Equal(t, expHeaders, record.Headers(path, http.MethodGet))
	assert.Equal(t, "documentPlain.pdf", u.FileName)
	assert.True(t, len(u.Payload) > 0)
}

func TestClientGet_Errors(t *testing.T) {
	path := "/upload/ID"

	// error http-request
	var record MockAssertion
	mockServer := NewMockServer(&record, MockServerProcedure{
		URI:        path,
		HTTPMethod: http.MethodGet,
		Response: MockResponse{
			StatusCode: 404,
		},
	})
	url := mockServer.URL
	client := upload.NewClient("/")
	_, err := client.Get("ID", "TOKEN")
	if err == nil {
		t.Errorf("error expected")
	}

	// no success
	record = MockAssertion{}
	mockServer = NewMockServer(&record, MockServerProcedure{
		URI:        path,
		HTTPMethod: http.MethodGet,
		Response: MockResponse{
			StatusCode: 404,
		},
	})
	url = mockServer.URL
	client = upload.NewClient(url)
	_, err = client.Get("ID", "TOKEN")
	if err == nil {
		t.Errorf("error expected")
	}

	// empty body
	record = MockAssertion{}
	mockServer = NewMockServer(&record, MockServerProcedure{
		URI:        path,
		HTTPMethod: http.MethodGet,
		Response: MockResponse{
			StatusCode: 200,
		},
	})
	url = mockServer.URL
	client = upload.NewClient(url)
	_, err = client.Get("ID", "TOKEN")
	if err == nil {
		t.Errorf("error expected")
	}

	// error http-request
	record = MockAssertion{}
	mockServer = NewMockServer(&record, MockServerProcedure{
		URI:        path,
		HTTPMethod: http.MethodGet,
		Response: MockResponse{
			StatusCode: 200,
			Body:       []byte(`{ body: }`),
		},
	})
	url = mockServer.URL
	client = upload.NewClient(url)
	_, err = client.Get("ID", "TOKEN")
	if err == nil {
		t.Errorf("error expected")
	}
}

func TestClientDELETE(t *testing.T) {
	path := "/upload/ID"
	var record MockAssertion
	mockServer := NewMockServer(&record, MockServerProcedure{
		URI:        path,
		HTTPMethod: http.MethodDelete,
		Response: MockResponse{
			StatusCode: 200,
		},
	})

	url := mockServer.URL
	client := upload.NewClient(url)
	err := client.Delete("ID", "TOKEN")
	if err != nil {
		t.Errorf("cannot get ID '%s', %v", "ID", err)
	}

	assert.Equal(t, 1, record.Hits(path, http.MethodDelete))
	// check request headers
	expHeaders := []http.Header{{
		"Authorization": []string{"Bearer TOKEN"},
		"Cache-Control": []string{"no-cache"},
	}}
	assert.Equal(t, expHeaders, record.Headers(path, http.MethodDelete))
}

func TestClientDELETE_Error(t *testing.T) {
	path := "/upload/ID"
	var record MockAssertion
	mockServer := NewMockServer(&record, MockServerProcedure{
		URI:        path,
		HTTPMethod: http.MethodDelete,
		Response: MockResponse{
			StatusCode: 500,
		},
	})

	url := mockServer.URL
	client := upload.NewClient(url)
	err := client.Delete("ID", "TOKEN")
	if err == nil {
		t.Errorf("error expected")
	}

	// no success
	record = MockAssertion{}
	mockServer = NewMockServer(&record, MockServerProcedure{
		URI:        path,
		HTTPMethod: http.MethodDelete,
		Response: MockResponse{
			StatusCode: 404,
		},
	})
	url = mockServer.URL
	client = upload.NewClient("/")
	_, err = client.Get("ID", "TOKEN")
	if err == nil {
		t.Errorf("error expected")
	}
}
