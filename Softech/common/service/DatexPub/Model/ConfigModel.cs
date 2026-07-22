using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace DatexPub.Model
{
	public class ConfigData
	{
		public DBInfo DBConfig { get; set; } = new();
		public string UrlAPIDatex { get; set; } = "";
		public string UrlEventiBolzano { get; set; } = "";
		public string PathPubblicazioni { get; set; } = "";
		public int TimeoutElaborazione { get; set; } = 0;
		public string PrefissoIdentificativi { get; set; } = "";
	}

	public class DBInfo
	{
		public string DBSource { get; set; } = string.Empty;
	}
}
