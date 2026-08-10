// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

using Microsoft.EntityFrameworkCore;
using NLog;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace DbManager.Utility
{
	public static partial class DbUtility
	{
		private static Logger logger = LogManager.GetLogger("DB");

		private static object lockDB = new();

		public static int SaveChanges(postgresContext DB)
		{
			try
			{
				lock (lockDB)
				{
					return DB.SaveChanges();
				}
			}
			catch (DbUpdateException ex)
			{
				logger.Error(ex.Message);
			}
			catch (Exception ex)
			{
				logger.Error(ex);
			}
			return -1;
		}
	}
}
