/*
Copyright 2019-2020 vChain, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tc

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/codenotary/immudb/pkg/client"
)

type immuTc struct {
	Client      client.ImmuClient
}

type ImmuTc interface {
	Start(context.Context) (err error)
}

func NewImmuTc(c client.ImmuClient) ImmuTc {
	return &immuTc{c}
}

func (s *immuTc) Start(ctx context.Context) (err error) {
	var r *schema.Root
	if r, err = s.Client.CurrentRoot(ctx); err != nil{
		return err
	}

	/*rand.Seed(time.Now().UnixNano())

	fmt.Printf("%v", r.Index)
	rand.Seed(time.Now().UnixNano())*/



	for i := uint64(0) ; i <= r.Index; i++{
		b := make([]byte, 8)
		if _, err := rand.Read(b);err != nil{
			return err
		}
		fmt.Println(binary.LittleEndian.Uint64(b))
	}

	return
}
