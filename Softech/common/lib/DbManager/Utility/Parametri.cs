using Supporto;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Reflection;
using System.Text;
using System.Threading.Tasks;

namespace DbManager.Utility
{
	public static partial class DbUtility
	{
		public static List<TAB_PARAMETRI> Parametri_LoadAll(postgresContext db)
		{
			try
			{
				return db.TAB_PARAMETRI.OrderBy(x => x.Modulo).ThenBy(x => x.Parametro).ToList();
			}
			catch (Exception ex)
			{
				LogUtility.LogException(logger, MethodBase.GetCurrentMethod(), ex);
				return new List<TAB_PARAMETRI>();
			}
		}

		public static TAB_PARAMETRI? Parametri_Load(string modulo, string parametro, postgresContext db)
		{
			try
			{
				return db.TAB_PARAMETRI.FirstOrDefault(x => x.Modulo == modulo && x.Parametro == parametro);
			}
			catch (Exception ex)
			{
				LogUtility.LogException(logger, MethodBase.GetCurrentMethod(), ex);
				return null;
			}
		}
	}
}
