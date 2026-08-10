// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Supporto
{
	public static class DateUtility
	{
		public static DateTime Conv(DateTime? dtmValore)
		{
			if (dtmValore != null)
				return Convert.ToDateTime(dtmValore);

			return DateTime.MinValue;
		}

		public static DateTime ConvLocal(DateTime? dtmValore)
		{
			if (dtmValore != null)
				return Convert.ToDateTime(dtmValore).ToLocalTime();

			return DateTime.MinValue;
		}

		public static DateTime? ConvNullableLocal(DateTime? dtmValore)
		{
			if (dtmValore != null)
				return Convert.ToDateTime(dtmValore).ToLocalTime();

			return null;
		}

		public static DateTime ConvUTC(DateTime? dtmValore)
		{
			if (dtmValore != null)
				return Convert.ToDateTime(dtmValore).ToUniversalTime();

			return DateTime.MinValue;
		}

		public static DateTime? ConvNullableUTC(DateTime? dtmValore)
		{
			if (dtmValore != null)
				return Convert.ToDateTime(dtmValore).ToUniversalTime();

			return null;
		}

		public static double Conv(double? dblValore)
		{
			if (dblValore != null)
				return Convert.ToDouble(dblValore);

			return 0;
		}

		public static decimal Conv(int? intValore)
		{
			if (intValore != null)
				return Convert.ToDecimal(intValore);

			return 0;
		}

		public static bool Conv(bool? blnValore)
		{
			if (blnValore != null)
				return Convert.ToBoolean(blnValore);

			return false;
		}

		public static int ConvNullable(int? intValore)
		{
			if (intValore != null)
				return Convert.ToInt32(intValore);

			return 0;
		}

		public static DateTime? ReadDateTimeLocal(DateTime? dtmValore)
		{
			if (dtmValore == null)
				return null;

			if (dtmValore.Value.Kind == DateTimeKind.Unspecified) return Conv(dtmValore).ToLocalTime();
			else if (dtmValore.Value.Kind == DateTimeKind.Utc) return Conv(dtmValore).ToLocalTime();
			else if (dtmValore.Value.Kind == DateTimeKind.Local) return dtmValore;

			return null;
		}

		public static DateTime ReadDateTimeLocal(DateTime dtmValore)
		{
			if (dtmValore.Kind == DateTimeKind.Unspecified) return dtmValore.ToLocalTime();
			else if (dtmValore.Kind == DateTimeKind.Utc) return dtmValore.ToLocalTime();
			else if (dtmValore.Kind == DateTimeKind.Local) return dtmValore;

			return DateTime.MinValue;
		}

		public static DateTime? ReadDateTimeUTC(DateTime? dtmValore)
		{
			if (dtmValore == null)
				return null;

			if (dtmValore.Value.Kind == DateTimeKind.Unspecified) return DateTime.SpecifyKind(dtmValore.Value, DateTimeKind.Utc);
			else if (dtmValore.Value.Kind == DateTimeKind.Utc) return dtmValore;
			else if (dtmValore.Value.Kind == DateTimeKind.Local) return ConvNullableUTC(dtmValore);

			return null;
		}

		public static DateTime ReadDateTimeUTC(DateTime dtmValore)
		{
			if (dtmValore.Kind == DateTimeKind.Unspecified) return DateTime.SpecifyKind(dtmValore, DateTimeKind.Utc);
			else if (dtmValore.Kind == DateTimeKind.Utc) return dtmValore;
			else if (dtmValore.Kind == DateTimeKind.Local) return ConvUTC(dtmValore);

			return DateTime.MinValue;
		}
	}
}
