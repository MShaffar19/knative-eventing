/*
Copyright 2020 The Knative Authors

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

package v1alpha1

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	eventingduckv1beta1 "knative.dev/eventing/pkg/apis/duck/v1beta1"
	"knative.dev/eventing/pkg/apis/flows/v1beta1"
	messagingv1beta1 "knative.dev/eventing/pkg/apis/messaging/v1beta1"
)

func TestSequenceConversionBadType(t *testing.T) {
	good, bad := &Sequence{}, &Sequence{}

	if err := good.ConvertUp(context.Background(), bad); err == nil {
		t.Errorf("ConvertUp() = %#v, wanted error", bad)
	}

	if err := good.ConvertDown(context.Background(), bad); err == nil {
		t.Errorf("ConvertDown() = %#v, wanted error", good)
	}
}

// Test v1alpha1 -> v1beta1 -> v1alpha1
func TestSequenceRoundTripV1alpha1(t *testing.T) {
	versions := []apis.Convertible{&v1beta1.Sequence{}}

	linear := eventingduckv1beta1.BackoffPolicyLinear

	tests := []struct {
		name string
		in   *Sequence
	}{{
		name: "min configuration",
		in: &Sequence{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "seq-name",
				Namespace:  "seq-ns",
				Generation: 17,
			},
			Spec: SequenceSpec{
				Steps: []SequenceStep{},
			},
		},
	}, {
		name: "full configuration",
		in: &Sequence{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "seq-name",
				Namespace:  "seq-ns",
				Generation: 17,
			},
			Spec: SequenceSpec{
				Steps: []SequenceStep{
					{
						Destination: duckv1.Destination{
							Ref: &duckv1.KReference{
								Kind:       "s1Kind",
								Namespace:  "s1Namespace",
								Name:       "s1Name",
								APIVersion: "s1APIVersion",
							},
							URI: apis.HTTP("s1.example.com")},
						Delivery: &eventingduckv1beta1.DeliverySpec{
							DeadLetterSink: &duckv1.Destination{
								Ref: &duckv1.KReference{
									Kind:       "dl1Kind",
									Namespace:  "dl1Namespace",
									Name:       "dl1Name",
									APIVersion: "dl1APIVersion",
								},
								URI: apis.HTTP("subscriber.dls1.example.com"),
							},
							Retry:         pointer.Int32Ptr(5),
							BackoffPolicy: &linear,
							BackoffDelay:  pointer.StringPtr("5s"),
						},
					},
					{
						Destination: duckv1.Destination{
							Ref: &duckv1.KReference{
								Kind:       "s2Kind",
								Namespace:  "s2Namespace",
								Name:       "s2Name",
								APIVersion: "s2APIVersion",
							},
							URI: apis.HTTP("s2.example.com")},
						Delivery: &eventingduckv1beta1.DeliverySpec{
							DeadLetterSink: &duckv1.Destination{
								Ref: &duckv1.KReference{
									Kind:       "dl2Kind",
									Namespace:  "dl2Namespace",
									Name:       "dl2Name",
									APIVersion: "dl2APIVersion",
								},
								URI: apis.HTTP("subscriber.dls2.example.com"),
							},
							Retry:         pointer.Int32Ptr(7),
							BackoffPolicy: &linear,
							BackoffDelay:  pointer.StringPtr("8s"),
						},
					},
				},
				ChannelTemplate: &messagingv1beta1.ChannelTemplateSpec{
					TypeMeta: metav1.TypeMeta{
						Kind:       "channelKind",
						APIVersion: "channelAPIVersion",
					},
				},
				Reply: &duckv1.Destination{
					Ref: &duckv1.KReference{
						Kind:       "replyKind",
						Namespace:  "replyNamespace",
						Name:       "replyName",
						APIVersion: "replyAPIVersion",
					},
					URI: apis.HTTP("reply.example.com"),
				},
			},
			Status: SequenceStatus{
				Status: duckv1.Status{
					ObservedGeneration: 1,
					Conditions: duckv1.Conditions{{
						Type:   "Ready",
						Status: "True",
					}},
				},
				AddressStatus: duckv1.AddressStatus{
					Address: &duckv1.Addressable{
						URL: apis.HTTP("addressstatus.example.com"),
					},
				},
				SubscriptionStatuses: []SequenceSubscriptionStatus{
					{
						Subscription: corev1.ObjectReference{
							Kind:       "s1-sub-kind",
							APIVersion: "s1-sub-apiversion",
							Name:       "s1-sub-name",
							Namespace:  "s1-sub-namespace",
						},
						ReadyCondition: apis.Condition{Message: "s1-msg"},
					},
					{
						Subscription: corev1.ObjectReference{
							Kind:       "s2-sub-kind",
							APIVersion: "s2-sub-apiversion",
							Name:       "s2-sub-name",
							Namespace:  "s2-sub-namespace",
						},
						ReadyCondition: apis.Condition{Message: "s2-msg"},
					},
				},
				ChannelStatuses: []SequenceChannelStatus{
					{
						Channel: corev1.ObjectReference{
							Kind:       "s1-channel-kind",
							APIVersion: "s1-channel-apiversion",
							Name:       "s1-channel-name",
							Namespace:  "s1-channel-namespace",
						},
						ReadyCondition: apis.Condition{Message: "s1-msg"},
					},
					{
						Channel: corev1.ObjectReference{
							Kind:       "s2-channel-kind",
							APIVersion: "s2-channel-apiversion",
							Name:       "s2-channel-name",
							Namespace:  "s2-channel-namespace",
						},
						ReadyCondition: apis.Condition{Message: "s2-msg"},
					},
				},
			},
		},
	}}

	for _, test := range tests {
		for _, version := range versions {
			t.Run(test.name, func(t *testing.T) {
				ver := version
				if err := test.in.ConvertUp(context.Background(), ver); err != nil {
					t.Errorf("ConvertUp() = %v", err)
				}
				got := &Sequence{}
				if err := got.ConvertDown(context.Background(), ver); err != nil {
					t.Errorf("ConvertDown() = %v", err)
				}

				if diff := cmp.Diff(test.in, got); diff != "" {
					t.Errorf("roundtrip (-want, +got) = %v", diff)
				}
			})
		}
	}
}

// Test v1beta1 -> v1alpha1 -> v1beta1
func TestSequenceRoundTripV1beta1(t *testing.T) {
	versions := []apis.Convertible{&Sequence{}}

	linear := eventingduckv1beta1.BackoffPolicyLinear

	tests := []struct {
		name string
		in   *v1beta1.Sequence
	}{{
		name: "min configuration",
		in: &v1beta1.Sequence{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "seq-name",
				Namespace:  "seq-ns",
				Generation: 17,
			},
			Spec: v1beta1.SequenceSpec{
				Steps: []v1beta1.SequenceStep{},
			},
		},
	}, {
		name: "full configuration",
		in: &v1beta1.Sequence{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "seq-name",
				Namespace:  "seq-ns",
				Generation: 17,
			},
			Spec: v1beta1.SequenceSpec{
				Steps: []v1beta1.SequenceStep{
					{
						Destination: duckv1.Destination{
							Ref: &duckv1.KReference{
								Kind:       "s1Kind",
								Namespace:  "s1Namespace",
								Name:       "s1Name",
								APIVersion: "s1APIVersion",
							},
							URI: apis.HTTP("s1.example.com")},
						Delivery: &eventingduckv1beta1.DeliverySpec{
							DeadLetterSink: &duckv1.Destination{
								Ref: &duckv1.KReference{
									Kind:       "dl1Kind",
									Namespace:  "dl1Namespace",
									Name:       "dl1Name",
									APIVersion: "dl1APIVersion",
								},
								URI: apis.HTTP("subscriber.dls1.example.com"),
							},
							Retry:         pointer.Int32Ptr(5),
							BackoffPolicy: &linear,
							BackoffDelay:  pointer.StringPtr("5s"),
						},
					},
					{
						Destination: duckv1.Destination{
							Ref: &duckv1.KReference{
								Kind:       "s2Kind",
								Namespace:  "s2Namespace",
								Name:       "s2Name",
								APIVersion: "s2APIVersion",
							},
							URI: apis.HTTP("s2.example.com")},
						Delivery: &eventingduckv1beta1.DeliverySpec{
							DeadLetterSink: &duckv1.Destination{
								Ref: &duckv1.KReference{
									Kind:       "dl2Kind",
									Namespace:  "dl2Namespace",
									Name:       "dl2Name",
									APIVersion: "dl2APIVersion",
								},
								URI: apis.HTTP("subscriber.dls2.example.com"),
							},
							Retry:         pointer.Int32Ptr(7),
							BackoffPolicy: &linear,
							BackoffDelay:  pointer.StringPtr("8s"),
						},
					},
				},
				ChannelTemplate: &messagingv1beta1.ChannelTemplateSpec{
					TypeMeta: metav1.TypeMeta{
						Kind:       "channelKind",
						APIVersion: "channelAPIVersion",
					},
				},
				Reply: &duckv1.Destination{
					Ref: &duckv1.KReference{
						Kind:       "replyKind",
						Namespace:  "replyNamespace",
						Name:       "replyName",
						APIVersion: "replyAPIVersion",
					},
					URI: apis.HTTP("reply.example.com"),
				},
			},
			Status: v1beta1.SequenceStatus{
				Status: duckv1.Status{
					ObservedGeneration: 1,
					Conditions: duckv1.Conditions{{
						Type:   "Ready",
						Status: "True",
					}},
				},
				AddressStatus: duckv1.AddressStatus{
					Address: &duckv1.Addressable{
						URL: apis.HTTP("addressstatus.example.com"),
					},
				},
				SubscriptionStatuses: []v1beta1.SequenceSubscriptionStatus{
					{
						Subscription: corev1.ObjectReference{
							Kind:       "s1-sub-kind",
							APIVersion: "s1-sub-apiversion",
							Name:       "s1-sub-name",
							Namespace:  "s1-sub-namespace",
						},
						ReadyCondition: apis.Condition{Message: "s1-msg"},
					},
					{
						Subscription: corev1.ObjectReference{
							Kind:       "s2-sub-kind",
							APIVersion: "s2-sub-apiversion",
							Name:       "s2-sub-name",
							Namespace:  "s2-sub-namespace",
						},
						ReadyCondition: apis.Condition{Message: "s2-msg"},
					},
				},
				ChannelStatuses: []v1beta1.SequenceChannelStatus{
					{
						Channel: corev1.ObjectReference{
							Kind:       "s1-channel-kind",
							APIVersion: "s1-channel-apiversion",
							Name:       "s1-channel-name",
							Namespace:  "s1-channel-namespace",
						},
						ReadyCondition: apis.Condition{Message: "s1-msg"},
					},
					{
						Channel: corev1.ObjectReference{
							Kind:       "s2-channel-kind",
							APIVersion: "s2-channel-apiversion",
							Name:       "s2-channel-name",
							Namespace:  "s2-channel-namespace",
						},
						ReadyCondition: apis.Condition{Message: "s2-msg"},
					},
				},
			},
		},
	}}

	for _, test := range tests {
		for _, version := range versions {
			t.Run(test.name, func(t *testing.T) {
				ver := version
				if err := ver.ConvertDown(context.Background(), test.in); err != nil {
					t.Errorf("ConvertDown() = %v", err)
				}
				got := &v1beta1.Sequence{}
				if err := ver.ConvertUp(context.Background(), got); err != nil {
					t.Errorf("ConvertDown() = %v", err)
				}

				if diff := cmp.Diff(test.in, got); diff != "" {
					t.Errorf("roundtrip (-want, +got) = %v", diff)
				}
			})
		}
	}
}