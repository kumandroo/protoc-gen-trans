package main

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/nutmeglabs/testify/assert"
	"github.com/pborman/uuid"

	"github.com/nutmeglabs/banda/gen/idl/extensions/protoc-gen-trans/test"
	"github.com/nutmeglabs/banda/libs/testutils"
)

type englishKeyGetter struct{}

func buildKey(s string) string {
	return uuid.NewSHA1(uuid.NIL, []byte(s)).String()
}

func (e *englishKeyGetter) GetKey(oldKey, translatedText string) string {
	return buildKey(translatedText)
}

type japaneseKeyGetter struct{}

func (e *japaneseKeyGetter) GetKey(oldKey, translatedText string) string {
	if oldKey != "" {
		return oldKey
	}

	return buildKey(translatedText)
}

func TestMessage1(t *testing.T) {
	msgEN := getEnglishMessage1()
	msgJP := getJapaneseMessage()
	msgENWithKeys := proto.Clone(msgEN).(*test.TestMessage1)

	enTranslations := msgENWithKeys.ExtractTranslations(nil, &englishKeyGetter{})

	msgJPWithKeys := proto.Clone(msgJP).(*test.TestMessage1)

	jpTranslations := msgJPWithKeys.ExtractTranslations(msgENWithKeys, &japaneseKeyGetter{})

	enKeys := msgENWithKeys.GetTranslationKeys()
	jpKeys := msgJPWithKeys.GetTranslationKeys()

	assert.Equal(t, enKeys, jpKeys)

	newMsgEN := proto.Clone(msgJPWithKeys).(*test.TestMessage1)
	newMsgEN.Translate(func(key string) string {
		return enTranslations[key]
	})

	testutils.AssertObjectsEqual(t, msgEN, newMsgEN)

	newMsgJP := proto.Clone(msgENWithKeys).(*test.TestMessage1)
	newMsgJP.Translate(func(key string) string {
		return jpTranslations[key]
	})

	testutils.AssertObjectsEqual(t, newMsgJP, msgJP)
}

func TestSecondaryLanguage(t *testing.T) {
	msgJP := getJapaneseMessage()
	msgJPWithKeys := proto.Clone(msgJP).(*test.TestMessage1)
	jpTranslations := msgJPWithKeys.ExtractTranslations(msgJPWithKeys, &japaneseKeyGetter{})

	newMsgJP := proto.Clone(msgJPWithKeys).(*test.TestMessage1)
	newMsgJP.Translate(func(key string) string {
		return jpTranslations[key]
	})

	assert.Equal(t, newMsgJP, msgJP)
}

func TestReuse(t *testing.T) {
	msg := &test.TestMessage1{
		Name1: "Hello, Goodbye",
	}
	_ = msg.ExtractTranslations(nil, &englishKeyGetter{})

	msgJP := &test.TestMessage1{
		Name1: "こんにちは、さよなら",
	}
	jpTranslations := msgJP.ExtractTranslations(msg, &japaneseKeyGetter{})

	msg2 := &test.TestMessage1{
		Name3: "Hello, Goodbye",
	}

	msg2.ExtractTranslations(nil, &englishKeyGetter{})

	msg2.Translate(func(key string) string {
		return jpTranslations[key]
	})

	assert.Equal(t, "こんにちは、さよなら", msg2.Name3)
}

// An array of empty strings in the source language should be overrideable by non-source language translations.
func TestEmptyStrings(t *testing.T) {
	msg1 := &test.TestMessage1{
		Array1: []string{
			"",
			"",
			"",
		},
	}
	_ = msg1.ExtractTranslations(nil, &englishKeyGetter{})

	oldMsg2 := &test.TestMessage1{
		Array1: []string{
			"abc",
			"def",
			"ghi",
		},
	}

	msg2WithKeys := proto.Clone(oldMsg2).(*test.TestMessage1)
	translations2 := msg2WithKeys.ExtractTranslations(msg1, &japaneseKeyGetter{})
	newMsg2 := proto.Clone(msg2WithKeys).(*test.TestMessage1)
	newMsg2.Translate(func(key string) string {
		return translations2[key]
	})

	assert.Equal(t, oldMsg2, newMsg2)
}

func getEnglishMessage1() *test.TestMessage1 {
	return &test.TestMessage1{
		Id:    uuid.NewSHA1(uuid.NameSpace_URL, []byte("1")).String(),
		Name1: "book",
		Name2: "CONSTANT",
		Name3: "person",
		Array1: []string{
			"one",
			"two",
			"three",
			"four",
		},

		Msg1: &test.TestMessage2{
			Name1: "movie",
			Name2: "CONSTANT",
			Name3: "dog",
			Array1: []string{
				"blue",
				"yellow",
				"green",
			},

			RecursiveMsgArray1: []*test.TestMessage2{
				{
					Name1: "table",
					Name2: "CONSTANT",
					Name3: "cat",
					Array1: []string{
						"happy",
						"sad",
						"mad",
					},
				},
				{
					Name1: "chair",
					Name2: "CONSTANT",
					Name3: "snake",
					Array1: []string{
						"finger",
						"hand",
						"arm",
					},
				},
			},
		},

		Msg2: &test.TestMessage1_NestedMessage{
			Name1: "backpack",
		},

		MessageMap: map[string]*test.TestMessage3{
			"1": &test.TestMessage3{
				Name1: "who",
				Name2: "CONSTANT",
			},
		},
	}
}

func getJapaneseMessage() *test.TestMessage1 {
	return &test.TestMessage1{
		Id:    uuid.NewSHA1(uuid.NameSpace_URL, []byte("1")).String(),
		Name1: "本",
		Name2: "CONSTANT",
		Name3: "人",
		Array1: []string{
			"一",
			"二",
			"三",
			"四",
		},

		Msg1: &test.TestMessage2{
			Name1: "映画",
			Name2: "CONSTANT",
			Name3: "犬",
			Array1: []string{
				"青い",
				"黄色い",
				"緑",
			},

			RecursiveMsgArray1: []*test.TestMessage2{
				{
					Name1: "テーブル",
					Name2: "CONSTANT",
					Name3: "猫",
					Array1: []string{
						"嬉しい",
						"悲しい",
						"怒っている",
					},
				},
				{
					Name1: "椅子",
					Name2: "CONSTANT",
					Name3: "蛇",
					Array1: []string{
						"指",
						"手",
						"腕",
					},
				},
			},
		},

		Msg2: &test.TestMessage1_NestedMessage{
			Name1: "リュックサック",
		},

		MessageMap: map[string]*test.TestMessage3{
			"1": &test.TestMessage3{
				Name1: "だれ",
				Name2: "CONSTANT",
			},
		},
	}
}

func getEnglishMessage2() *test.TestMessage1 {
	return &test.TestMessage1{
		Id:    uuid.NewSHA1(uuid.NameSpace_URL, []byte("1")).String(),
		Name1: "novela",
		Name2: "CONSTANT",
		Name3: "truck driver",
		Array1: []string{
			"tons",
			"lots",
			"loads",
			"craploads",
		},

		Msg1: &test.TestMessage2{
			Name1: "drama",
			Name2: "CONSTANT",
			Name3: "GOD",
			Array1: []string{
				"azure",
				"sunlight",
				"sea green",
			},

			RecursiveMsgArray1: []*test.TestMessage2{
				{
					Name1: "nightstand",
					Name2: "CONSTANT",
					Name3: "tabby",
					Array1: []string{
						"ecstatic",
						"devastated",
						"furious",
					},
					RecursiveMsgArray1: []*test.TestMessage2{
						{
							Name1: "cerulean",
						},
					},
				},
				{
					Name1: "cambered",
					Array1: []string{
						"truculent",
						"long-bodied",
						"neuralgic",
					},
				},
			},
		},

		Msg2: &test.TestMessage1_NestedMessage{
			Name1: "backpack",
		},
	}
}
