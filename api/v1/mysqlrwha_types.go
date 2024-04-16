/*
Copyright 2024.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MysqlrwhaSpec defines the desired state of Mysqlrwha
type Mysql struct {
	MysqlImage       string `json:"mysqlimage"`
	MysqlReplicas    int32  `json:"mysqlreplicas"`
	Xtrabackupimage  string `json:"xtrabackupimage"`
	MysqlMem         string `json:"mysqlmem"`
	MysqlCpu         string `json:"mysqlcpu"`
	Mysqlstoragename string `json:"mysqlstoragename"`
	Mysqlstoragemax  string `json:"mysqlstoragemax"`
	//mysql的root用户初始化密码
	MysqlRootPassword string `json:"mysqlrootpassword"`
	// mysql主从同步的用户及密码
	MysqlReplicaUser string `json:"mysqlreplicauser"`
	MysqlReplicapwd  string `json:"mysqlreplicapwd"`
	// mysql初始化数据库名
	Mysqldatabase string `json:"mysqldatabase"`
	// 需同步的数据库列表
	Replicadb []string `json:"replicadb"`
	// 忽略同步的数据库列表
	Ingnoredb []string `json:"ingnoredb,omitempty"`
}
type Mycat struct {
	Mycatimage    string `json:"mycatimage,omitempty"`
	MycatReplicas int32  `json:"mycatreplicas,omitempty"`
	MycatMem      string `json:"mycatmem"`
	MycatCpu      string `json:"mycatcpu"`
	// 访问mycat的用户密码
	MycatWriteUser string `json:"mysqlwriteuser"`
	MycatWritePwd  string `json:"mysqlwritepwd"`
	MycatReadUser  string `json:"mycatwriteuser"`
	MycatReadPwd   string `json:"mycatwritepwd"`
	//只读mycat逻辑库,逻辑库名不能大写
	Mycatreaddb []string `json:"mycatreaddb"`
	//读写mycat逻辑库,逻辑库名不能大写
	Mycatwritedb []string `json:"mycatwritedb"`
	// 访问mycat逻辑库(key)及与mysql对应的数据库,其中key要与读写的权限的逻辑库一致,逻辑库名不能大写
	Mycatrwdb map[string]string `json:"mycatrwdb"`
}
type MysqlrwhaSpec struct {
	Mysql           *Mysql `json:"mysql"`
	Mycat           *Mycat `json:"mycat"`
	ImagePullPolicy string `json:"imagepullpolicy"`
}

// MysqlrwhaStatus defines the observed state of Mysqlrwha
type MysqlrwhaStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Mysqlrwha is the Schema for the mysqlrwhas API
type Mysqlrwha struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MysqlrwhaSpec   `json:"spec,omitempty"`
	Status MysqlrwhaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MysqlrwhaList contains a list of Mysqlrwha
type MysqlrwhaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mysqlrwha `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Mysqlrwha{}, &MysqlrwhaList{})
}
