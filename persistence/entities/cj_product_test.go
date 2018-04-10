package entities

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_IsHelmet_returns_true_when_the_category_ends_with_Motorcycle_Helmets(t *testing.T) {
	RegisterTestingT(t)

	product := &CJProduct{Category: "jlasdkfjasdf Motorcycle Helmets", Name: "yolo"}
	Expect(product.IsHelmet()).To(BeTrue())

	product.Category = "Motorcycle Helmets"
	Expect(product.IsHelmet()).To(BeTrue())
}

func Test_IsHelmet_returns_false_when_the_category_ends_with_Motorcycle_Helmets_but_Name_contains_shield(t *testing.T) {
	RegisterTestingT(t)

	product := &CJProduct{Category: "jlasdkfjasdf Motorcycle Helmets", Name: "Shoei RF-LEGIT  ShIeLD"}
	Expect(product.IsHelmet()).To(BeFalse())
}

func Test_IsHelmet_returns_false_when_the_category_ends_with_Motorcycle_Helmets_but_Name_contains_spoiler(t *testing.T) {
	RegisterTestingT(t)

	product := &CJProduct{Category: "jlasdkfjasdf Motorcycle Helmets", Name: "Shoei RF-LEGIT SPOILER"}
	Expect(product.IsHelmet()).To(BeFalse())
}

func Test_IsHelmet_returns_false_when_the_category_does_not_end_with_Motorcycle_Helmets(t *testing.T) {
	RegisterTestingT(t)

	product := &CJProduct{Category: "jlasdkfjasdf Motorcycle Helmetz"}
	Expect(product.IsHelmet()).To(BeFalse())
}
