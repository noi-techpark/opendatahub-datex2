// SPDX-FileCopyrightText: 2026 2026 NOI Techpark <digital@noi.bz.it>
// SPDX-FileCopyrightText: 2026, 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: CC0-1.0

using Microsoft.Extensions.Logging;
using NLog;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Reflection;
using System.Text;
using System.Threading.Tasks;

namespace Supporto
{
	public static class LogUtility
	{
		public static void LogException(Logger logger, MethodBase? metodo, Exception ex)
		{
			if (metodo == null)
				return;

			logger.Error(metodo.Name + "> " + ex.Message);

			if (ex.InnerException != null)
			{
				logger.Error(metodo.Name + ".Inner> " + ex.InnerException.Message);

				if (ex.InnerException.InnerException != null)
				{
					logger.Error(metodo.Name + ".InnerBis> " + ex.InnerException.InnerException.Message);
				}
			}
		}
	}

	public enum RequestError
	{
		NONE = 0,
		GENERIC = 1,
		DB_READ = 2,
		DISABLED = 3,
		NOT_FOUND = 4,
		BAD_DATA = 5,
		NO_DATA = 6,
		MSG_FOR_USER = 7,
		BAD_SESSION = 8
	}

	public static class RequestErrorMessages
	{
		public static string GENERIC { get; } = "Errore generico nell'elaborazione";
		public static string DB_READ { get; } = "Errore nella lettura dei dati su DB";
		public static string DISABLED { get; } = "L'elemento ricercato è disabilitato";
		public static string NOT_FOUND { get; } = "L'elemento ricercato non è stato trovato";
		public static string BAD_DATA { get; } = "Parametri non validi";
		public static string BAD_SESSION { get; } = "IdSessione non valido";
		public static string NO_DATA { get; } = "Nessun dato trovato";
	}
}
