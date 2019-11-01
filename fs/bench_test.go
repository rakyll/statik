// Copyright 2019 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fs

import "testing"

func BenchmarkOpen(b *testing.B) {
	Register(mustZipTree("../testdata/index"))
	fs, err := New()
	if err != nil {
		b.Fatalf("New() = %v", err)
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			const name = "/index.html"
			_, err := fs.Open(name)
			if err != nil {
				b.Errorf("fs.Open(%v) = %v", name, err)
			}
		}
	})
}

func BenchmarkOpenDeep(b *testing.B) {
	Register(mustZipTree("../testdata/deep"))
	fs, err := New()
	if err != nil {
		b.Fatalf("New() = %v", err)
	}
	for i := 0; i < b.N; i++ {
		const name = "/aa/bb/c"
		_, err := fs.Open(name)
		if err != nil {
			b.Errorf("fs.Open(%v) = %v", name, err)
		}
	}
}
