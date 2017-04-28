package cluster

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/tanji/replication-manager/dbhelper"
)

func (cluster *Cluster) initMdbsproxy(oldmaster *ServerMonitor, proxy *Proxy) {
	params := fmt.Sprintf("?timeout=%ds", cluster.conf.Timeout)

	dsn := proxy.User + ":" + proxy.Pass + "@"
	dsn += "tcp(" + proxy.Host + ":" + proxy.Port + ")/" + params
	c, err := sqlx.Open("mysql", dsn)
	if err != nil {
		cluster.LogPrintf("ERROR: Could not connect to MariaDB Sharding Proxy %s", err)
		cluster.mdbsBootstrap(proxy)
	}
	err = c.Ping()
	if err != nil {
		cluster.LogPrintf("ERROR: Could not connect to MariaDB Sharding Proxy %s", err)
		cluster.mdbsBootstrap(proxy)
	}
	schemas, err := dbhelper.GetSchemas(cluster.master.Conn)
	if err != nil {
		cluster.LogPrintf("ERROR: Could not fetch master schemas %s", err)
	}

	for _, s := range schemas {
		c.Exec("CREATE OR REPLACE SERVER master_" + s + "_" + cluster.GetName() + " OPTIONS HOST '" + cluster.master.Host + "' DATABASE '" + s + "' USER '" + cluster.master.User + "' PASSWORD '" + cluster.master.Pass + "' PORT " + cluster.master.Port)
		c.Exec("CREATE DATABASE IS NOT EXISTS " + s)
	}
	tables, err := dbhelper.GetTables(cluster.master.Conn)
	if err != nil {
		cluster.LogPrintf("ERROR: Could not fetch master tables %s", err)
	}
	for _, t := range tables {
		c.Exec("CREATE OR REPLACE TABLE " + t.Table_schema + "." + t.Table_name + " ENGINE=SPIDER comment='srv=\"master_" + t.Table_schema + "_\"" + cluster.GetName())
	}
	c.Close()
}

func (cluster *Cluster) mdbsCreateTables(proxy *Proxy) {
}

func (cluster *Cluster) mdbsBootstrap(proxy *Proxy) {
	cluster.LogPrintf("INFO: Bootstrap MariaDB Sharding Cluster")
	srv, err := cluster.newServerMonitor(proxy.Host+":"+proxy.Port, proxy.User, proxy.Pass)
	srv.ClusterGroup = cluster
	if err != nil {
		cluster.LogPrintf("ERROR: Bootstrap MariaDB Sharding Cluster Failed")

		return
	}
	cluster.StartMariaDB(srv)
	sql := "INSERT INTO mysql.proc VALUES ('mysql','spider_plugin_installer','PROCEDURE','spider_plugin_installer','SQL','CONTAINS_SQL','NO','DEFINER','','',0x626567696E0A2020736574204077696E5F706C7567696E203A3D20494628404076657273696F6E5F636F6D70696C655F6F73206C696B65202757696E25272C20312C2030293B0A20207365742040686176655F7370696465725F706C7567696E203A3D20303B0A202073656C6563742040686176655F7370696465725F706C7567696E203A3D20312066726F6D20494E464F524D4154494F4E5F534348454D412E706C7567696E7320776865726520504C5547494E5F4E414D45203D2027535049444552273B0A202069662040686176655F7370696465725F706C7567696E203D2030207468656E200A202020206966204077696E5F706C7567696E203D2030207468656E200A202020202020696E7374616C6C20706C7567696E2073706964657220736F6E616D65202768615F7370696465722E736F273B0A20202020656C73650A202020202020696E7374616C6C20706C7567696E2073706964657220736F6E616D65202768615F7370696465722E646C6C273B0A20202020656E642069663B0A2020656E642069663B0A20207365742040686176655F7370696465725F695F735F616C6C6F635F6D656D5F706C7567696E203A3D20303B0A202073656C6563742040686176655F7370696465725F695F735F616C6C6F635F6D656D5F706C7567696E203A3D20312066726F6D20494E464F524D4154494F4E5F534348454D412E706C7567696E7320776865726520504C5547494E5F4E414D45203D20275350494445525F414C4C4F435F4D454D273B0A202069662040686176655F7370696465725F695F735F616C6C6F635F6D656D5F706C7567696E203D2030207468656E200A202020206966204077696E5F706C7567696E203D2030207468656E200A202020202020696E7374616C6C20706C7567696E207370696465725F616C6C6F635F6D656D20736F6E616D65202768615F7370696465722E736F273B0A20202020656C73650A202020202020696E7374616C6C20706C7567696E207370696465725F616C6C6F635F6D656D20736F6E616D65202768615F7370696465722E646C6C273B0A20202020656E642069663B0A2020656E642069663B0A20207365742040686176655F7370696465725F6469726563745F73716C5F756466203A3D20303B0A202073656C6563742040686176655F7370696465725F6469726563745F73716C5F756466203A3D20312066726F6D206D7973716C2E66756E63207768657265206E616D65203D20277370696465725F6469726563745F73716C273B0A202069662040686176655F7370696465725F6469726563745F73716C5F756466203D2030207468656E0A202020206966204077696E5F706C7567696E203D2030207468656E200A2020202020206372656174652066756E6374696F6E207370696465725F6469726563745F73716C2072657475726E7320696E7420736F6E616D65202768615F7370696465722E736F273B0A20202020656C73650A2020202020206372656174652066756E6374696F6E207370696465725F6469726563745F73716C2072657475726E7320696E7420736F6E616D65202768615F7370696465722E646C6C273B0A20202020656E642069663B0A2020656E642069663B0A20207365742040686176655F7370696465725F62675F6469726563745F73716C5F756466203A3D20303B0A202073656C6563742040686176655F7370696465725F62675F6469726563745F73716C5F756466203A3D20312066726F6D206D7973716C2E66756E63207768657265206E616D65203D20277370696465725F62675F6469726563745F73716C273B0A202069662040686176655F7370696465725F62675F6469726563745F73716C5F756466203D2030207468656E0A202020206966204077696E5F706C7567696E203D2030207468656E200A202020202020637265617465206167677265676174652066756E6374696F6E207370696465725F62675F6469726563745F73716C2072657475726E7320696E7420736F6E616D65202768615F7370696465722E736F273B0A20202020656C73650A202020202020637265617465206167677265676174652066756E6374696F6E207370696465725F62675F6469726563745F73716C2072657475726E7320696E7420736F6E616D65202768615F7370696465722E646C6C273B0A20202020656E642069663B0A2020656E642069663B0A20207365742040686176655F7370696465725F70696E675F7461626C655F756466203A3D20303B0A202073656C6563742040686176655F7370696465725F70696E675F7461626C655F756466203A3D20312066726F6D206D7973716C2E66756E63207768657265206E616D65203D20277370696465725F70696E675F7461626C65273B0A202069662040686176655F7370696465725F70696E675F7461626C655F756466203D2030207468656E0A202020206966204077696E5F706C7567696E203D2030207468656E200A2020202020206372656174652066756E6374696F6E207370696465725F70696E675F7461626C652072657475726E7320696E7420736F6E616D65202768615F7370696465722E736F273B0A20202020656C73650A2020202020206372656174652066756E6374696F6E207370696465725F70696E675F7461626C652072657475726E7320696E7420736F6E616D65202768615F7370696465722E646C6C273B0A20202020656E642069663B0A2020656E642069663B0A20207365742040686176655F7370696465725F636F70795F7461626C65735F756466203A3D20303B0A202073656C6563742040686176655F7370696465725F636F70795F7461626C65735F756466203A3D20312066726F6D206D7973716C2E66756E63207768657265206E616D65203D20277370696465725F636F70795F7461626C6573273B0A202069662040686176655F7370696465725F636F70795F7461626C65735F756466203D2030207468656E0A202020206966204077696E5F706C7567696E203D2030207468656E200A2020202020206372656174652066756E6374696F6E207370696465725F636F70795F7461626C65732072657475726E7320696E7420736F6E616D65202768615F7370696465722E736F273B0A20202020656C73650A2020202020206372656174652066756E6374696F6E207370696465725F636F70795F7461626C65732072657475726E7320696E7420736F6E616D65202768615F7370696465722E646C6C273B0A20202020656E642069663B0A2020656E642069663B0A0A20207365742040686176655F7370696465725F666C7573685F7461626C655F6D6F6E5F63616368655F756466203A3D20303B0A202073656C6563742040686176655F7370696465725F666C7573685F7461626C655F6D6F6E5F63616368655F756466203A3D20312066726F6D206D7973716C2E66756E63207768657265206E616D65203D20277370696465725F666C7573685F7461626C655F6D6F6E5F6361636865273B0A202069662040686176655F7370696465725F666C7573685F7461626C655F6D6F6E5F63616368655F756466203D2030207468656E0A202020206966204077696E5F706C7567696E203D2030207468656E200A2020202020206372656174652066756E6374696F6E207370696465725F666C7573685F7461626C655F6D6F6E5F63616368652072657475726E7320696E7420736F6E616D65202768615F7370696465722E736F273B0A20202020656C73650A2020202020206372656174652066756E6374696F6E207370696465725F666C7573685F7461626C655F6D6F6E5F63616368652072657475726E7320696E7420736F6E616D65202768615F7370696465722E646C6C273B0A20202020656E642069663B0A2020656E642069663B0A0A656E64,'root@localhost','2014-11-07 03:42:40','2014-11-07 03:42:40','','','utf8','utf8_general_ci','latin1_swedish_ci',0x626567696E0A2020736574204077696E5F706C7567696E203A3D20494628404076657273696F6E5F636F6D70696C655F6F73206C696B65202757696E25272C20312C2030293B0A20207365742040686176655F7370696465725F706C7567696E203A3D20303B0A202073656C6563742040686176655F7370696465725F706C7567696E203A3D20312066726F6D20494E464F524D4154494F4E5F534348454D412E706C7567696E7320776865726520504C5547494E5F4E414D45203D2027535049444552273B0A202069662040686176655F7370696465725F706C7567696E203D2030207468656E200A202020206966204077696E5F706C7567696E203D2030207468656E200A202020202020696E7374616C6C20706C7567696E2073706964657220736F6E616D65202768615F7370696465722E736F273B0A20202020656C73650A202020202020696E7374616C6C20706C7567696E2073706964657220736F6E616D65202768615F7370696465722E646C6C273B0A20202020656E642069663B0A2020656E642069663B0A20207365742040686176655F7370696465725F695F735F616C6C6F635F6D656D5F706C7567696E203A3D20303B0A202073656C6563742040686176655F7370696465725F695F735F616C6C6F635F6D656D5F706C7567696E203A3D20312066726F6D20494E464F524D4154494F4E5F534348454D412E706C7567696E7320776865726520504C5547494E5F4E414D45203D20275350494445525F414C4C4F435F4D454D273B0A202069662040686176655F7370696465725F695F735F616C6C6F635F6D656D5F706C7567696E203D2030207468656E200A202020206966204077696E5F706C7567696E203D2030207468656E200A202020202020696E7374616C6C20706C7567696E207370696465725F616C6C6F635F6D656D20736F6E616D65202768615F7370696465722E736F273B0A20202020656C73650A202020202020696E7374616C6C20706C7567696E207370696465725F616C6C6F635F6D656D20736F6E616D65202768615F7370696465722E646C6C273B0A20202020656E642069663B0A2020656E642069663B0A20207365742040686176655F7370696465725F6469726563745F73716C5F756466203A3D20303B0A202073656C6563742040686176655F7370696465725F6469726563745F73716C5F756466203A3D20312066726F6D206D7973716C2E66756E63207768657265206E616D65203D20277370696465725F6469726563745F73716C273B0A202069662040686176655F7370696465725F6469726563745F73716C5F756466203D2030207468656E0A202020206966204077696E5F706C7567696E203D2030207468656E200A2020202020206372656174652066756E6374696F6E207370696465725F6469726563745F73716C2072657475726E7320696E7420736F6E616D65202768615F7370696465722E736F273B0A20202020656C73650A2020202020206372656174652066756E6374696F6E207370696465725F6469726563745F73716C2072657475726E7320696E7420736F6E616D65202768615F7370696465722E646C6C273B0A20202020656E642069663B0A2020656E642069663B0A20207365742040686176655F7370696465725F62675F6469726563745F73716C5F756466203A3D20303B0A202073656C6563742040686176655F7370696465725F62675F6469726563745F73716C5F756466203A3D20312066726F6D206D7973716C2E66756E63207768657265206E616D65203D20277370696465725F62675F6469726563745F73716C273B0A202069662040686176655F7370696465725F62675F6469726563745F73716C5F756466203D2030207468656E0A202020206966204077696E5F706C7567696E203D2030207468656E200A202020202020637265617465206167677265676174652066756E6374696F6E207370696465725F62675F6469726563745F73716C2072657475726E7320696E7420736F6E616D65202768615F7370696465722E736F273B0A20202020656C73650A202020202020637265617465206167677265676174652066756E6374696F6E207370696465725F62675F6469726563745F73716C2072657475726E7320696E7420736F6E616D65202768615F7370696465722E646C6C273B0A20202020656E642069663B0A2020656E642069663B0A20207365742040686176655F7370696465725F70696E675F7461626C655F756466203A3D20303B0A202073656C6563742040686176655F7370696465725F70696E675F7461626C655F756466203A3D20312066726F6D206D7973716C2E66756E63207768657265206E616D65203D20277370696465725F70696E675F7461626C65273B0A202069662040686176655F7370696465725F70696E675F7461626C655F756466203D2030207468656E0A202020206966204077696E5F706C7567696E203D2030207468656E200A2020202020206372656174652066756E6374696F6E207370696465725F70696E675F7461626C652072657475726E7320696E7420736F6E616D65202768615F7370696465722E736F273B0A20202020656C73650A2020202020206372656174652066756E6374696F6E207370696465725F70696E675F7461626C652072657475726E7320696E7420736F6E616D65202768615F7370696465722E646C6C273B0A20202020656E642069663B0A2020656E642069663B0A20207365742040686176655F7370696465725F636F70795F7461626C65735F756466203A3D20303B0A202073656C6563742040686176655F7370696465725F636F70795F7461626C65735F756466203A3D20312066726F6D206D7973716C2E66756E63207768657265206E616D65203D20277370696465725F636F70795F7461626C6573273B0A202069662040686176655F7370696465725F636F70795F7461626C65735F756466203D2030207468656E0A202020206966204077696E5F706C7567696E203D2030207468656E200A2020202020206372656174652066756E6374696F6E207370696465725F636F70795F7461626C65732072657475726E7320696E7420736F6E616D65202768615F7370696465722E736F273B0A20202020656C73650A2020202020206372656174652066756E6374696F6E207370696465725F636F70795F7461626C65732072657475726E7320696E7420736F6E616D65202768615F7370696465722E646C6C273B0A20202020656E642069663B0A2020656E642069663B0A0A20207365742040686176655F7370696465725F666C7573685F7461626C655F6D6F6E5F63616368655F756466203A3D20303B0A202073656C6563742040686176655F7370696465725F666C7573685F7461626C655F6D6F6E5F63616368655F756466203A3D20312066726F6D206D7973716C2E66756E63207768657265206E616D65203D20277370696465725F666C7573685F7461626C655F6D6F6E5F6361636865273B0A202069662040686176655F7370696465725F666C7573685F7461626C655F6D6F6E5F63616368655F756466203D2030207468656E0A202020206966204077696E5F706C7567696E203D2030207468656E200A2020202020206372656174652066756E6374696F6E207370696465725F666C7573685F7461626C655F6D6F6E5F63616368652072657475726E7320696E7420736F6E616D65202768615F7370696465722E736F273B0A20202020656C73650A2020202020206372656174652066756E6374696F6E207370696465725F666C7573685F7461626C655F6D6F6E5F63616368652072657475726E7320696E7420736F6E616D65202768615F7370696465722E646C6C273B0A20202020656E642069663B0A2020656E642069663B0A0A656E64)"
	srv.Conn.Exec(sql)
	srv.Conn.Exec("CALL mysql.spider_plugin_installer")
	srv.Close()
}
