package provisioning

import (
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cstorapis "github.com/openebs/api/pkg/apis/cstor/v1"
	"github.com/openebs/cstor-operators/tests/pkg/cspc/cspcspecbuilder"
	"github.com/openebs/cstor-operators/tests/pkg/cstorvolumeconfig/cvcspecbuilder"
	"github.com/openebs/cstor-operators/tests/pkg/k8sclient"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	cstorCSIProvisionerName = "cstor.csi.openebs.io"
)

// ProvisionCSIVolume will provision Volume using CStor-CSI
func ProvisionCSIVolume(pvcName, pvcNamespace, scName string) {
	var (
		// PVC will contains the PersistentVolumeClaim object created for test
		pvc *corev1.PersistentVolumeClaim
		// SC will contains the StorageClass object created for test
		sc *storagev1.StorageClass
	)

	parameters := map[string]string{
		"cas-type":         "cstor",
		"cstorPoolCluster": cspc.Name,
		"replicaCount":     strconv.Itoa(cstorsuite.ReplicaCount),
	}
	sc = createStorageClass(scName, parameters)

	err := cstorsuite.client.CreateNamespace(pvcNamespace)
	Expect(err).To(BeNil())

	pvc = createPersistentVolumeClaim(pvcName, pvcNamespace, sc.Name)

	err = cstorsuite.client.WaitForPersistentVolumeClaimPhase(
		pvc.Name, pvc.Namespace, corev1.ClaimBound, k8sclient.Poll, k8sclient.ClaimBindingTimeout)
	Expect(err).To(BeNil())

}

// createStorageClass in etcd
func createStorageClass(scName string, parameters map[string]string) *storagev1.StorageClass {
	// Building StorageClass
	var err error
	var isVolumeExpansionAllowed bool
	sc := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: scName,
		},
		Provisioner:          cstorCSIProvisionerName,
		AllowVolumeExpansion: &isVolumeExpansionAllowed,
		Parameters:           parameters,
	}

	By(fmt.Sprintf("Creating %s StorageClass", sc.Name))

	sc, err = cstorsuite.client.KubeClientSet.StorageV1().StorageClasses().Create(sc)
	Expect(err).To(BeNil())
	return sc
}

func createPersistentVolumeClaim(pvcName, pvcNamespace, scName string) *corev1.PersistentVolumeClaim {
	var err error
	resCapacity, err := resource.ParseQuantity("5G")
	Expect(err).To(BeNil())

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: pvcNamespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &scName,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): resCapacity,
				},
			},
		},
	}

	pvc, err = cstorsuite.client.KubeClientSet.
		CoreV1().
		PersistentVolumeClaims(pvcNamespace).
		Create(pvc)
	Expect(err).To(BeNil())

	return pvc
}

// ProvisionCSPC will create CSPC based cStor pools
func ProvisionCSPC(cspcName, namespace, poolType string, bdCount int) {
	cspcSpecBuilder = cspcspecbuilder.
		NewCSPCSpecBuilder(cstorsuite.CSPCCache, cstorsuite.infra)

	cspc = cspcSpecBuilder.BuildCSPC(cspcName, namespace, poolType, bdCount, cstorsuite.infra.NodeCount).GetCSPCSpec()
	_, err := cstorsuite.
		client.
		OpenEBSClientSet.
		CstorV1().
		CStorPoolClusters(cspc.Namespace).
		Create(cspc)
	Expect(err).To(BeNil())

	gotHealthyCSPiCount := cstorsuite.
		client.
		GetOnlineCSPICountEventually(cspc.Name, cspc.Namespace, cstorsuite.infra.NodeCount)
	Expect(gotHealthyCSPiCount).To(BeNumerically("==", int32(cstorsuite.infra.NodeCount)))

	poolList, err := cstorsuite.client.GetCStorPoolInstanceNames(cspc.Name, cspc.Namespace)
	Expect(err).To(BeNil())

	//Intantiating CVCSpecBuilder to reduce etcd calls later
	cvcSpecBuilder = cvcspecbuilder.NewCVCSpecBuilder(cstorsuite.infra, poolList)
}

// DeProvisionCSPC will de-provision CSPC based cStor pools
func DeProvisionCSPC(cspc *cstorapis.CStorPoolCluster) {
	err := cstorsuite.
		client.
		OpenEBSClientSet.
		CstorV1().
		CStorPoolClusters(cspc.Namespace).
		Delete(cspc.Name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil())

	cspcSpecBuilder.ResetCSPCSpecData()

	gotCSPICount := cstorsuite.
		client.
		GetCSPICountEventually(cspc.Name, cspc.Namespace, 0)
	Expect(gotCSPICount).To(BeNumerically("==", 0))
}

// DeProvisionVolume will delete the provided PVC from system
func DeProvisionVolume(pvcName, pvcNamespace, scName string) {
	// Read the PVC after before deleting the PVC
	pvc, err := cstorsuite.client.KubeClientSet.
		CoreV1().
		PersistentVolumeClaims(pvcNamespace).
		Get(pvcName, metav1.GetOptions{})
	if err != nil && !k8serror.IsNotFound(err) {
		Expect(err).To(BeNil())
	}

	err = cstorsuite.client.KubeClientSet.
		CoreV1().
		PersistentVolumeClaims(pvcNamespace).
		Delete(pvcName, &metav1.DeleteOptions{})
	if err != nil && !k8serror.IsNotFound(err) {
		Expect(err).To(BeNil())
	}

	err = cstorsuite.client.WaitForPersistentVolumeClaimDeletion(pvcName, pvcNamespace, k8sclient.Poll, k8sclient.ClaimDeletingTimeout)
	Expect(err).To(BeNil())


	err = cstorsuite.client.KubeClientSet.
		StorageV1().
		StorageClasses().
		Delete(scName, &metav1.DeleteOptions{})
	if err != nil && !k8serror.IsNotFound(err) {
		Expect(err).To(BeNil())
	}

	if pvc.Spec.VolumeName == "" {
		klog.Errorf("PVC %s in %s namespace is not in Bound to PV", pvcName, pvcNamespace)
		return
	}

	err = cstorsuite.client.WaitForCStorVolumeDeletion(
		pvc.Spec.VolumeName, openebsNamespace, k8sclient.Poll, k8sclient.CVDeletingTimeout)
	Expect(err).To(BeNil())

	err = cstorsuite.client.WaitForVolumeManagerCountEventually(
		pvc.Spec.VolumeName, openebsNamespace, 0, k8sclient.Poll, k8sclient.CVCPhaseTimeout)
	Expect(err).To(BeNil())

	err = cstorsuite.client.WaitForCVRCountEventually(
		pvc.Spec.VolumeName, openebsNamespace, 0,
		k8sclient.Poll, k8sclient.CVRPhaseTimeout)
	Expect(err).To(BeNil())

	err = cstorsuite.client.WaitForCStorVolumeConfigDeletion(
		pvc.Spec.VolumeName, openebsNamespace, k8sclient.Poll, k8sclient.CVCDeletingTimeout)
	Expect(err).To(BeNil())

	// Might be case where CVC is not marked as bound
	if cvcSpecBuilder.CVC != nil {
		cvcSpecBuilder.UnsetCVCSpec()
	}
}

// VerifyCStorVolumeResourcesStatus will verifies the CStorVolume resources health state
func VerifyCStorVolumeResourcesStatus(pvcName, pvcNamespace string) {
	err := cstorsuite.client.WaitForPersistentVolumeClaimPhase(
		pvcName, pvcNamespace, corev1.ClaimBound, k8sclient.Poll, k8sclient.ClaimBindingTimeout)
	Expect(err).To(BeNil())

	// Read the PVC after bind so that it will contain pv name
	pvc, err := cstorsuite.client.KubeClientSet.
		CoreV1().
		PersistentVolumeClaims(pvcNamespace).
		Get(pvcName, metav1.GetOptions{})
	Expect(err).To(BeNil())

	err = cstorsuite.client.WaitForCStorVolumeConfigPhase(
		pvc.Spec.VolumeName, openebsNamespace, cstorapis.CStorVolumeConfigPhaseBound, k8sclient.Poll, k8sclient.CVCPhaseTimeout)
	Expect(err).To(BeNil())

	err = cstorsuite.client.WaitForVolumeManagerCountEventually(
		pvc.Spec.VolumeName, openebsNamespace, 1, k8sclient.Poll, k8sclient.CVCPhaseTimeout)
	Expect(err).To(BeNil())

	err = cstorsuite.client.WaitForCStorVolumePhase(
		pvc.Spec.VolumeName, openebsNamespace, cstorapis.CStorVolumePhase("Healthy"), k8sclient.Poll, k8sclient.CVPhaseTimeout)
	Expect(err).To(BeNil())

	err = cstorsuite.client.WaitForCVRCountEventually(
		pvc.Spec.VolumeName, openebsNamespace, cstorsuite.ReplicaCount,
		k8sclient.Poll, k8sclient.CVRPhaseTimeout, cstorapis.IsCVRHealthy)
	Expect(err).To(BeNil())

	cvc, err := cstorsuite.client.GetCVC(pvc.Spec.VolumeName, openebsNamespace)
	Expect(err).To(BeNil())

	Expect(cvcSpecBuilder).NotTo(BeNil(), "cvcSpecBuilder is not initilized")
	// SetCVCSpec will hold CVC object inmemory to perform further actions
	cvcSpecBuilder.SetCVCSpec(cvc)
}
