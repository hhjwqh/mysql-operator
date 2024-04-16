package utils

import (
	"bytes"
	"github.com/hhjwqh/mysql-operator/api/v1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"strconv"

	"strings"
	"text/template"
)

func parseTemplate(templateName string, app *v1.Mysqlrwha) []byte {
	tmpl, err := template.ParseFiles("controllers/template/" + templateName + ".yaml")
	if err != nil {
		panic(err)
	}
	b := new(bytes.Buffer)
	err = tmpl.Execute(b, app)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func NewStatefullset(app *v1.Mysqlrwha) *appv1.StatefulSet {
	s := &appv1.StatefulSet{}
	err := yaml.Unmarshal(parseTemplate("mysqlstatefulset", app), s)
	if err != nil {
		panic(err)
	}
	return s
}

func NewHeadlessService(app *v1.Mysqlrwha) *corev1.Service {
	hs := &corev1.Service{}
	err := yaml.Unmarshal(parseTemplate("headlessservice", app), hs)
	if err != nil {
		panic(err)
	}
	return hs
}
func NewMysqlConfigmap(app *v1.Mysqlrwha) *corev1.ConfigMap {
	c := &corev1.ConfigMap{}
	c.ObjectMeta.Name = app.ObjectMeta.Name + "-mysql-configmap"
	c.ObjectMeta.Namespace = app.ObjectMeta.Namespace

	m := make(map[string]string)
	primary1 := `
	[mysqld]
	datadir=/var/lib/mysql
	socket=/var/lib/mysql/mysql.sock
	symbolic-links=0
	pid-file=/var/run/mysqld/mysqld.pid
	binlog_format=row
	log-bin
	relay-log = relay-log
    `
	var dstr string
	dstr = ""
	if len(app.Spec.Mysql.Replicadb) > 0 {
		for _, v := range app.Spec.Mysql.Replicadb {
			dstr = dstr + "\tbinlog-do-db=" + v + "\n"
		}
	}
	var istr string
	istr = ""
	if len(app.Spec.Mysql.Ingnoredb) > 0 {
		for _, v := range app.Spec.Mysql.Ingnoredb {
			istr = istr + "\tbinlog-ignore-db=" + v + "\n"
		}
	}
	primary2 := `
	default-storage-engine = INNODB
	lower_case_table_names=1
	gtid_mode=ON
	enforce-gtid-consistency=ON
	log-slave-updates=1
	rpl_semi_sync_master_enabled=1
	plugin-load=rpl_semi_sync_master=semisync_master.so
	rpl_semi_sync_master_timeout=1000
	binlog_group_commit_sync_delay=100
	binlog_group_commit_sync_no_delay_count=10
	sync_binlog=1
	innodb_flush_log_at_trx_commit=1
	max_connections = 400
	max_connect_errors = 1000
	explicit_defaults_for_timestamp = true
	interactive_timeout = 1800
	wait_timeout = 1800
	max_allowed_packet = 1024M
	tmp_table_size = 134217728
	max_heap_table_size = 134217728
	query_cache_size = 0
	query_cache_type = 0
	read_buffer_size = 131072
	sort_buffer_size = 2097152
	binlog_cache_size = 524288
	expire_logs_days = 30
   `

	replica1 := `
	[mysqld]
    datadir=/var/lib/mysql
    socket=/var/lib/mysql/mysql.sock
    symbolic-links=0
    pid-file=/var/run/mysqld/mysqld.pid
    relay-log=relay-log
    log-bin
    read_only=1
    relay_log_purge=0
    log_slave_updates=1
    `
	replica2 := `
	default-storage-engine=MyISAM
    lower_case_table_names=1
    binlog_format=row
    gtid_mode=ON
    enforce-gtid-consistency=ON
    slave-parallel-type=LOGICAL_CLOCK
    slave-parallel-workers=4
    master_info_repository=TABLE
    relay_log_info_repository=TABLE
    relay_log_recovery=ON
    plugin-load=rpl_semi_sync_slave=semisync_slave.so
    rpl_semi_sync_slave_enabled=1
    max_connections = 400
    max_connect_errors = 1000
    explicit_defaults_for_timestamp = true 
    interactive_timeout = 1800
    wait_timeout = 1800
    max_allowed_packet = 1024M  
    tmp_table_size = 134217728
    max_heap_table_size = 134217728
    query_cache_size = 0
    query_cache_type = 0
    read_buffer_size = 131072
    sort_buffer_size = 2097152
    binlog_cache_size = 524288  
    expire_logs_days = 30
	`
	//primary1 + dstr + istr + primary2
	m["primary.cnf"] = primary1 + dstr + istr + primary2
	m["replica.cnf"] = replica1 + dstr + istr + replica2
	//c.Data["replica.cnf"] = replica1 + dstr + istr + replica2
	c.Data = m

	return c
}

func NewMycatConfigmap(app *v1.Mysqlrwha) *corev1.ConfigMap {
	c := &corev1.ConfigMap{}
	c.ObjectMeta.Name = app.ObjectMeta.Name + "-mycat-configmap"
	c.ObjectMeta.Namespace = app.ObjectMeta.Namespace

	m := make(map[string]string)
	server := `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mycat:server SYSTEM "server.dtd">
<mycat:server xmlns:mycat="http://io.mycat/">
		<system>
		<property name="nonePasswordLogin">0</property> <!-- 0为需要密码登陆、1为不需要密码登陆 ,默认为0，设置为1则需要指定默认账户-->
		<property name="useHandshakeV10">1</property>
		<property name="useSqlStat">0</property>  <!-- 1为开启实时统计、0为关闭 -->
		<property name="useGlobleTableCheck">0</property>  <!-- 1为开启全加班一致性检测、0为关闭 -->
		<property name="sequnceHandlerType">2</property>
		<property name="subqueryRelationshipCheck">false</property> <!-- 子查询中存在关联查询的情况下,检查关联字段中是否有分片字段 .默认 false -->
				<property name="processorBufferPoolType">0</property>
				<property name="handleDistributedTransactions">0</property>
				<property name="useOffHeapForMerge">1</property>
				<property name="memoryPageSize">64k</property>
				<property name="spillsFileBufferSize">1k</property>
				<property name="useStreamOutput">0</property>
				<property name="systemReserveMemorySize">384m</property>
				<property name="useZKSwitch">false</property>
				<property name="strictTxIsolation">false</property>
				<property name="useZKSwitch">true</property>
		</system>
		<user name="WRITEUSER" >
				<property name="password">WRITEPWD</property>
				<property name="schemas">WRITEDBS</property>
		</user>
		<user name="READUSER">
				<property name="password">READPWD</property>
				<property name="schemas">READDBS</property>
				<property name="readOnly">true</property>
		</user>
</mycat:server>
	`
	writedbs := strings.Join(app.Spec.Mycat.Mycatwritedb, ",")
	readdbs := strings.Join(app.Spec.Mycat.Mycatreaddb, ",")
	server = strings.Replace(server, "WRITEDBS", writedbs, -1)
	server = strings.Replace(server, "READDBS", readdbs, -1)
	server = strings.Replace(server, "WRITEUSER", app.Spec.Mycat.MycatWriteUser, -1)
	server = strings.Replace(server, "WRITEPWD", app.Spec.Mycat.MycatWritePwd, -1)
	server = strings.Replace(server, "READUSER", app.Spec.Mycat.MycatReadUser, -1)
	server = strings.Replace(server, "READPWD", app.Spec.Mycat.MycatReadPwd, -1)

	m["server.xml"] = server

	schema1 := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mycat:schema SYSTEM "schema.dtd">
<mycat:schema xmlns:mycat="http://io.mycat/">`
	mycatdbs := ""
	if len(app.Spec.Mycat.Mycatrwdb) > 0 {
		n := 1
		for key, v := range app.Spec.Mycat.Mycatrwdb {
			mycatstr := `     <schema name="MYCATDB" checkSQLschema="false" sqlMaxLimit="100" dataNode="DATENODEDDD">
        </schema>
	   <dataNode name="DATENODEDDD" dataHost="MYCATLOCALHOST" database="MYSQLDATABASE" />
	   <dataHost name="MYCATLOCALHOST" maxCon="200" minCon="20" balance="3" writeType="0" dbType="mysql" dbDriver="native" switchType="1"  slaveThreshold="100">
	           <heartbeat>select user()</heartbeat>
	           <writeHost host="hostM0" url="MYSQLHOST:3306" user="root" password="MysqlRootPassword">
                   READHOST
	           </writeHost>
	   </dataHost>`
			readstr := `<readHost host="HOSTS" url="READDBHOST:3306" user="root" password="MysqlRootPassword" />`
			ndstr := "dn" + strconv.Itoa(n)
			host := "localhost" + strconv.Itoa(n)
			readstrs := ""
			mysqlhost0str := app.Name + "-mysql-0." + app.Name + "-headless"

			if app.Spec.Mysql.MysqlReplicas > 1 {
				for i := 1; i < int(app.Spec.Mysql.MysqlReplicas); i++ {
					hosts := "hostM" + strconv.Itoa(i)
					urlstr := app.Name + "-mysql-" + strconv.Itoa(i) + "." + app.Name + "-headless"
					read := strings.Replace(readstr, "MysqlRootPassword", app.Spec.Mysql.MysqlRootPassword, -1)
					read = strings.Replace(read, "HOSTS", hosts, -1)
					read = strings.Replace(read, "READDBHOST", urlstr, -1)
					if i+1 >= int(app.Spec.Mysql.MysqlReplicas) {
						readstrs = readstrs + read
					} else {
						readstrs = readstrs + read + "\n                 "
					}
				}
			}
			mycatstr = strings.Replace(mycatstr, "MysqlRootPassword", app.Spec.Mysql.MysqlRootPassword, -1)
			mycatstr = strings.Replace(mycatstr, "MYCATDB", key, -1)
			mycatstr = strings.Replace(mycatstr, "DATENODEDDD", ndstr, -1)
			mycatstr = strings.Replace(mycatstr, "MYCATLOCALHOST", host, -1)
			mycatstr = strings.Replace(mycatstr, "MYSQLDATABASE", v, -1)
			mycatstr = strings.Replace(mycatstr, "MYSQLHOST", mysqlhost0str, -1)
			mycatstr = strings.Replace(mycatstr, "READHOST", readstrs, -1)
			if n+1 >= len(app.Spec.Mycat.Mycatrwdb) {
				mycatdbs = mycatdbs + mycatstr
			} else {
				mycatdbs = mycatdbs + mycatstr + "\n              "
			}
			n = n + 1
		}

	}
	m["schema.xml"] = schema1 + "\n" + mycatdbs + "\n</mycat:schema>"

	c.Data = m
	//// 将数据结构转换为 YAML 格式
	//yamlData, err := y.Marshal(&c)
	//if err != nil {
	//	log.Fatalf("error: %v", err)
	//}
	//
	//// 将 YAML 数据写入文件
	//err = os.WriteFile("controllers/template/mycaoconfig.yaml", yamlData, 0644)
	//if err != nil {
	//	log.Fatalf("error: %v", err)
	//}
	return c
}

func NewMycatConfigmapb(app *v1.Mysqlrwha) *corev1.ConfigMap {
	c := &corev1.ConfigMap{}
	c.ObjectMeta.Name = app.ObjectMeta.Name + "-mycat-configmap"
	c.ObjectMeta.Namespace = app.ObjectMeta.Namespace
	err := yaml.Unmarshal(parseTemplate("mycatconfigmap", app), c)
	if err != nil {
		panic(err)
	}

	return c
}
func NewMycatService(app *v1.Mysqlrwha) *corev1.Service {
	s := &corev1.Service{}
	err := yaml.Unmarshal(parseTemplate("mycatservice", app), s)
	if err != nil {
		panic(err)
	}
	return s
}
func NewMycatDeploy(app *v1.Mysqlrwha) *appv1.Deployment {
	d := &appv1.Deployment{}
	err := yaml.Unmarshal(parseTemplate("mycatdeployment", app), d)
	if err != nil {
		panic(err)
	}
	return d
}
