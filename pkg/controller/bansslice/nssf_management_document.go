package bansslice

import (
	"context"
	bansv1alpha1 "github.com/stevenchiu30801/bans5gc-operator/pkg/apis/bans/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NsiInformation struct {
	NrfId string `json:"nrfId"`
	NsiId string `json:"nsiId,omitempty"`
}

// NssfManagementItem defines the single management object of NSSF
type NssfManagementItem struct {
	SnssaiList []bansv1alpha1.Snssai `json:"snssaiList"`

	PlmnIdList []bansv1alpha1.PlmnId `json:"plmnIdList"`

	TaiList []bansv1alpha1.Tai `json:"taiList"`

	NsiInformationList []NsiInformation `json:"nsiInformationList"`
}

type NssfManagementDocument []NssfManagementItem

// newNssfManagementDocument returns a new NssfManagementDocument object to configure NSSF
func (r *ReconcileBansSlice) newNssfManagementDocument(b *bansv1alpha1.BansSlice) (NssfManagementDocument, error) {
	// Generate NSI information
	nrfList := &corev1.PodList{}
	opts := []client.ListOption{
		client.InNamespace(b.Namespace),
		client.MatchingLabels(map[string]string{"app.kubernetes.io/instance": "free5gc", "app.kubernetes.io/name": "nrf"}),
	}
	err := r.client.List(context.TODO(), nrfList, opts...)
	if err != nil {
		return NssfManagementDocument{}, err
	}
	// Access the first NRF
	nrfId := "https://" + nrfList.Items[0].Status.PodIP + ":29510"

	return NssfManagementDocument{
		{
			SnssaiList: b.Spec.SnssaiList,
			PlmnIdList: []bansv1alpha1.PlmnId{b.Spec.Tai.PlmnId},
			TaiList:    []bansv1alpha1.Tai{b.Spec.Tai},
			NsiInformationList: []NsiInformation{
				{
					NrfId: nrfId,
				},
			},
		},
	}, nil
}
