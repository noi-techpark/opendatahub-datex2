// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

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
		public static List<TAB_SOTTOTIPI> Sottotipi_LoadAll(postgresContext db)
		{
			try
			{
				return db.TAB_SOTTOTIPI.OrderBy(x => x.IdSottotipo).ToList();
			}
			catch (Exception ex)
			{
				LogUtility.LogException(logger, MethodBase.GetCurrentMethod(), ex);
				return new List<TAB_SOTTOTIPI>();
			}
		}
	}
}
