package resources

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"operator/mysql/pkg/apis/app/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MySQLService(mysql *v1.MySQL) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mysql.Name,
			Namespace: mysql.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mysql, schema.GroupVersionKind{
					Group:   v1.SchemeGroupVersion.Group,
					Version: v1.SchemeGroupVersion.Version,
					Kind:    "MySQL",
				}),
			},
		},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeNodePort,
			Ports: mysql.Spec.Ports,
			Selector: map[string]string{
				"app": mysql.Name,
			},
		},
	}
}
