// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging;
using NLog;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Reflection;
using System.Text;
using System.Threading.Tasks;

namespace DbManager
{
	public class DbConfiguration
	{
		public static string DBSource { get; set; } = string.Empty;

		public static bool SetConfiguration(string dbsource)
		{
			DBSource = dbsource;
			//LegacyDBHelper.logger.Info("DBSource: " + LegacyDbConfiguration.DBSource);

			return true;
		}
	}

	public static class DBHelper
	{
		public static Logger logger = LogManager.GetLogger("DB");

		public static postgresContext? Connect()
		{
			try
			{
				//logger.Debug("DBSource: " + DbConfiguration.DBSource);
				DbContextOptions<postgresContext> ctx = new DbContextOptionsBuilder<postgresContext>().UseNpgsql(DbConfiguration.DBSource).Options;
				return new postgresContext(ctx);
			}
			catch (Exception ex)
			{
				logger.Error(string.Format("____{0}) - Message: {1}. Inner Exception:{2}", MethodBase.GetCurrentMethod()?.Name ?? "Connect", ex.Message, ex.InnerException == null ? "" : ex.InnerException.Message));
				return null;
			}
		}
	}
}
