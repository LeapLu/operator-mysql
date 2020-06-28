package resources

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	appv1 "operator/mysql/pkg/apis/app/v1"
	"operator/mysql/pkg/apis/app/v1"
)

func MySQLDeployment(mysql *appv1.MySQL) *appsv1.Deployment {
	labels := map[string]string{
		"app": mysql.Name,
	}
	selecor := &metav1.LabelSelector{
		MatchLabels: labels,
	}
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
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
		Spec: appsv1.DeploymentSpec{
			Replicas: mysql.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: MySQLContainers(mysql),
				},
			},
			Selector: selecor,
		},
	}

}

func MySQLContainers(mysql *appv1.MySQL) []corev1.Container {
	containerPorts := []corev1.ContainerPort{}
	envVar := []corev1.EnvVar{}
	env := corev1.EnvVar{}
	env.Name = "MYSQL_ROOT_PASSWORD"
	env.Value = mysql.Spec.RootPassword
	envVar = append(envVar, env)
	for _, svcPort := range mysql.Spec.Ports {
		cport := corev1.ContainerPort{}
		cport.ContainerPort = svcPort.TargetPort.IntVal
		containerPorts = append(containerPorts, cport)
	}
	return []corev1.Container{
		{
			Name:            mysql.Name,
			Image:           mysql.Spec.Image,
			Ports:           containerPorts,
			Env:             envVar,
			ImagePullPolicy: corev1.PullIfNotPresent,
		},
	}
}
