// Copyright 2020-2021 Buf Technologies, Inc.
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

package bufimagemodify

import (
	"context"
	"strings"

	"github.com/bufbuild/buf/internal/buf/bufimage"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// rubyPackagePath is the SourceCodeInfo path for the ruby_package option.
// https://github.com/protocolbuffers/protobuf/blob/61689226c0e3ec88287eaed66164614d9c4f2bf7/src/google/protobuf/descriptor.proto#L453
var rubyPackagePath = []int32{8, 45}

func rubyPackage(sweeper Sweeper) Modifier {
	return ModifierFunc(
		func(ctx context.Context, image bufimage.Image) error {
			for _, imageFile := range image.Files() {
				if err := rubyPackageForFile(ctx, sweeper, imageFile); err != nil {
					return err
				}
			}
			return nil
		},
	)
}

func rubyPackageForFile(
	ctx context.Context,
	sweeper Sweeper,
	imageFile bufimage.ImageFile,
) error {
	descriptor := imageFile.Proto()
	rubyPackageValue := rubyPackageValue(imageFile)
	if isWellKnownType(ctx, imageFile) || rubyPackageValue == "" {
		// This is a well-known type or we could not resolve a non-empty ruby_package
		// value, so this is a no-op.
		return nil
	}
	if descriptor.Options == nil {
		descriptor.Options = &descriptorpb.FileOptions{}
	}
	descriptor.Options.RubyPackage = proto.String(rubyPackageValue)
	if sweeper != nil {
		sweeper.mark(imageFile.Path(), rubyPackagePath)
	}
	return nil
}

// rubyPackageValue returns the ruby_package for the given ImageFile based on its
// package declaration. If the image file doesn't have a package declaration, an
// empty string is returned.
func rubyPackageValue(imageFile bufimage.ImageFile) string {
	pkg := imageFile.Proto().GetPackage()
	if pkg == "" {
		return ""
	}
	packageParts := strings.Split(pkg, ".")
	for i, part := range packageParts {
		packageParts[i] = strings.Title(part)
	}
	return strings.Join(packageParts, "::")
}