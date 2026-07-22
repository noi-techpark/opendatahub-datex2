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
		public static List<TAB_DESTINATARI> Destinatari_LoadAll(postgresContext db)
		{
			try
			{
				return db.TAB_DESTINATARI.OrderBy(x => x.IdDestinatario).ToList();
			}
			catch (Exception ex)
			{
				LogUtility.LogException(logger, MethodBase.GetCurrentMethod(), ex);
				return new List<TAB_DESTINATARI>();
			}
		}
	}
}
