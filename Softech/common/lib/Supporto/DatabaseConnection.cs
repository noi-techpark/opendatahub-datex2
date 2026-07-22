using Npgsql;
using System;
using System.Collections;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Supporto
{
	public class DatabaseConnection
	{
		#region DATA
		private string m_strStringConn = "";
		private string m_strLastError = "";
		private NpgsqlConnection? m_conPgs = null;
		private NpgsqlTransaction? m_trnPgs = null;
		private static Dictionary<string, DataTable> m_tableMem = new Dictionary<string, DataTable>();
		#endregion

		public DatabaseConnection(string strStringConn)
		{
			try
			{
				m_strStringConn = strStringConn;

				m_conPgs = new NpgsqlConnection(strStringConn);
				m_conPgs.Open();

				m_strLastError = "";
			}
			catch (Exception ex)
			{
				m_strLastError = ex.Message;
				throw new Exception(ex.Message);
			}
		}

		public string GetLastError()
		{
			return m_strLastError;
		}

		public bool Close()
		{
			try
			{
				if (m_conPgs != null)
				{
					m_conPgs.Dispose();
					m_conPgs.Close();
				}

				m_strLastError = "";
				return true;
			}
			catch (Exception ex)
			{
				m_strLastError = ex.Message;
				return false;
			}
		}

		public bool BeginTrans()
		{
			try
			{
				if (m_conPgs == null)
					return false;

				m_trnPgs = m_conPgs.BeginTransaction(IsolationLevel.ReadCommitted);
				m_strLastError = "";
				return true;
			}
			catch (Exception ex)
			{
				m_strLastError = ex.Message;
				return false;
			}
		}

		public bool CommitTrans()
		{
			try
			{
				if (m_conPgs == null)
					return false;

				if (m_trnPgs != null)
				{ 
					m_trnPgs.Commit();

					m_trnPgs.Dispose();
					m_trnPgs = null;
				}

				m_strLastError = "";
				return true;
			}
			catch (Exception ex)
			{
				m_strLastError = ex.Message;
				return false;
			}
		}

		public bool RollbackTrans()
		{
			try
			{
				if (m_trnPgs != null)
				{
					m_trnPgs.Rollback();

					m_trnPgs.Dispose();
					m_trnPgs = null;
				}

				m_strLastError = "";
				return true;
			}
			catch (Exception ex)
			{
				m_strLastError = ex.Message;
				return false;
			}
		}

		public bool TestConnection(string strQuery)
		{
			try
			{
				if (m_conPgs == null)
					return false;

				DataTable? tab = null;

				if (m_conPgs.State != ConnectionState.Open)
					m_conPgs.Open();

				tab = GetTable("test", strQuery);

				if (tab == null)
				{
					m_strLastError = "Errore durante TestConnection con query: " + strQuery;
					return false;
				}

				m_strLastError = "";
				return true;
			}
			catch (Exception ex)
			{
				m_strLastError = ex.Message;
				return false;
			}
		}

		public DataTable? GetTableMem(string strTableName, string strQuery)
		{
			try
			{
				lock (m_tableMem)
				{
					DataTable? mem = null;
					if (m_tableMem.TryGetValue(strQuery, out mem))
						return mem;

					DataTable? tab = GetTable(strTableName, strQuery);
					if (tab != null)
						m_tableMem.Add(strQuery, tab);

					// mantengo un massimo di query
					if (m_tableMem.Count > 1000)
						m_tableMem.Remove(m_tableMem.Keys.First());

					return tab;
				}
			}
			catch (Exception ex)
			{
				m_strLastError = ex.Message;
				return null;
			}
		}

		public DataTable? GetTable(string strTableName, string strQuery)
		{
			try
			{
				if (m_conPgs == null)
					return null;

				// la tabella deve avere un nome per poter essere serializzata
				if (strTableName == "")
					strTableName = "Table";

				DataTable? tabDati = null;

				if (m_conPgs.State != ConnectionState.Open)
					m_conPgs.Open();

				NpgsqlDataReader dbReader = GetPgsDataReader(m_conPgs, m_trnPgs, strQuery);
				tabDati = new DataTable(strTableName);
				tabDati.Load(dbReader);

				m_strLastError = "";
				return tabDati;
			}
			catch (Exception ex)
			{
				m_strLastError = ex.Message;
				return null;
			}
		}

		private NpgsqlDataReader GetPgsDataReader(NpgsqlConnection conConnessione, NpgsqlTransaction? trnTransazione, string strQuery)
		{
			if (conConnessione.State != ConnectionState.Open)
				conConnessione.Open();

			NpgsqlCommand command = new NpgsqlCommand();
			command.Connection = m_conPgs;
			command.CommandType = CommandType.Text;
			command.Transaction = trnTransazione;
			command.CommandText = strQuery;

			return command.ExecuteReader();
		}

		public bool Execute(string strSql)
		{
			try
			{
				if (m_conPgs == null)
					return false;

				if (m_conPgs.State != ConnectionState.Open)
					m_conPgs.Open();

				NpgsqlCommand command = new NpgsqlCommand();
				command.Connection = m_conPgs;
				command.CommandType = CommandType.Text;
				command.Transaction = m_trnPgs;
				command.CommandText = strSql;

				command.ExecuteNonQuery();

				m_strLastError = "";
				return true;
			}
			catch (Exception ex)
			{
				m_strLastError = ex.Message;
				return false;
			}
		}

		public object ExecuteAndReturnID(string strQuery)
		{
			try
			{
				if (m_conPgs == null)
					return false;

				if (m_conPgs.State != ConnectionState.Open)
					m_conPgs.Open();

				NpgsqlCommand objCommand = new NpgsqlCommand();
				objCommand.Connection = m_conPgs;
				objCommand.CommandText = strQuery;

				NpgsqlParameter objParam = new NpgsqlParameter(":ID", NpgsqlTypes.NpgsqlDbType.Integer);
				objParam.Direction = ParameterDirection.Output;
				objCommand.Parameters.Add(objParam);
				objCommand.Transaction = m_trnPgs;

				objCommand.ExecuteNonQuery();

				m_strLastError = "";
				if (objParam.Value == null)
					return false;

				return objParam.Value;
			}
			catch (Exception ex)
			{
				m_strLastError = ex.Message;
				return false;
			}
		}

		public object? ExecuteScalar(string strSql)
		{
			try
			{
				if (m_conPgs == null)
					return false;

				if (m_conPgs.State != ConnectionState.Open)
					m_conPgs.Open();

				NpgsqlCommand command = new NpgsqlCommand();
				command.Connection = m_conPgs;
				command.CommandType = CommandType.Text;
				command.Transaction = m_trnPgs;
				command.CommandText = strSql;

				m_strLastError = "";
				return command.ExecuteScalar();
			}
			catch (Exception ex)
			{
				m_strLastError = ex.Message;
				return false;
			}
		}

		public DataTable DistinctDt(DataTable tabDt, string strColumn)
		{
			DataTable tabDistinct = tabDt.Clone();
			Hashtable htKey = new Hashtable();
			foreach (DataRow row in tabDt.Rows)
			{
				if (!htKey.Contains(row[strColumn]))
				{
					tabDistinct.ImportRow(row);
					htKey.Add(row[strColumn], null);
				}
			}
			return tabDistinct;
		}

		public static string? GetValueString(DataRow rowDati, string strCampo)
		{
			try
			{
				if (rowDati[strCampo] == DBNull.Value)
					return "";

				return Convert.ToString(rowDati[strCampo]);
			}
			catch (Exception ex)
			{
				string err = ex.Message;
				return "";
			}
		}

		public static int GetValueInt(DataRow rowDati, string strCampo)
		{
			try
			{
				if (rowDati[strCampo] == DBNull.Value)
					return 0;

				return Convert.ToInt32(rowDati[strCampo]);
			}
			catch (Exception ex)
			{
				string err = ex.Message;
				return 0;
			}
		}

		public static double GetValueDouble(DataRow rowDati, string strCampo)
		{
			try
			{
				if (rowDati[strCampo] == DBNull.Value)
					return 0.0;

				return Convert.ToDouble(rowDati[strCampo]);
			}
			catch (Exception ex)
			{
				string err = ex.Message;
				return 0.0;
			}
		}

		public static decimal GetValueDecimal(DataRow rowDati, string strCampo)
		{
			try
			{
				if (rowDati[strCampo] == DBNull.Value)
					return 0;

				return Convert.ToDecimal(rowDati[strCampo]);
			}
			catch (Exception ex)
			{
				string err = ex.Message;
				return 0;
			}
		}

		public static DateTime GetValueDataTime(DataRow rowDati, string strCampo)
		{
			try
			{
				if (rowDati[strCampo] == DBNull.Value)
					return DateTime.MinValue;

				return Convert.ToDateTime(rowDati[strCampo]).ToLocalTime();
			}
			catch (Exception ex)
			{
				string err = ex.Message;
				return DateTime.MinValue;
			}
		}

		public static DateTime GetValueDataTimeUTC(DataRow rowDati, string strCampo)
		{
			try
			{
				if (rowDati[strCampo] == DBNull.Value)
					return DateTime.MinValue;

				return Convert.ToDateTime(rowDati[strCampo]);
			}
			catch (Exception ex)
			{
				string err = ex.Message;
				return DateTime.MinValue;
			}
		}

		public static bool GetValueBool(DataRow rowDati, string strCampo)
		{
			try
			{
				if (rowDati[strCampo] == DBNull.Value)
					return false;

				return Convert.ToBoolean(rowDati[strCampo]);
			}
			catch (Exception ex)
			{
				string err = ex.Message;
				return false;
			}
		}

		private static string GetTableCampi(DataTable tab)
		{
			string strCampi = "";
			foreach (DataColumn col in tab.Columns)
			{
				if (strCampi != "")
					strCampi += ",";

				strCampi += col.ColumnName;
			}

			return strCampi;
		}

		private static string GetTableTypes(DataTable tab)
		{
			string strType = "";
			foreach (DataColumn col in tab.Columns)
			{
				if (strType != "")
					strType += ",";

				strType += col.DataType.FullName;
			}

			return strType;
		}

		private static string GetRowScript(DataRow row, string strCampi, string[] strTypes)
		{
			if (row == null)
				return "";

			if (strCampi == "")
			{
				strCampi = GetTableCampi(row.Table);
				strTypes = GetTableTypes(row.Table).Split(',');
			}

			string strValori = "";
			for (int intCol = 0; intCol < strTypes.Length; intCol++)
			{
				if (strValori != "")
					strValori += ", ";

				object objDato = row[intCol];

				strValori += GetDatoValue(objDato, strTypes[intCol]);
			}

			return "INSERT INTO " + row.Table.TableName + " (" + strCampi + ") VALUES (" + strValori + ")";
		}

		private static string? GetDatoValue(object objDato, string strType)
		{
			if (objDato == DBNull.Value)
			{
				return "NULL";
			}
			else if (strType == "System.String" ||
					 strType == "System.Char")
			{
				return "'" + objDato.ToString().Replace("'", "''") + "'";
			}
			else if (strType == "System.Int16" ||
					 strType == "System.Int32" ||
					 strType == "System.Int64" ||
					 strType == "System.UInt16" ||
					 strType == "System.UInt32" ||
					 strType == "System.UInt64" ||
					 strType == "System.Byte" ||
					 strType == "System.SByte" ||
					 strType == "System.Single")
			{
				return objDato.ToString();
			}
			else if (strType == "System.Double" || strType == "System.Decimal")
			{
				return objDato.ToString().Replace(',', '.');
			}
			else if (strType == "System.Boolean")
			{
				if (Convert.ToBoolean(objDato) == true)
					return "1";
				else
					return "0";
			}
			else if (strType == "System.DateTime")
			{
				return "TO_DATE('" + ((DateTime)objDato).ToString("yyyy/MM/dd HH:mm:ss") + "', 'yyyy/MM/dd hh24:mi:ss')";
			}
			else if (strType == "System.TimeSpan")
			{
				TimeSpan tmsAppo = (TimeSpan)objDato;
				string strFormat = tmsAppo.Days + " " + tmsAppo.Hours + ":" + tmsAppo.Minutes + ":" + tmsAppo.Seconds;
				return "INTERVAL '" + strFormat + "' DAY TO SECOND";
			}

			return "NULL";
		}

		public static string GetRowScript(DataRow row)
		{
			if (row == null)
				return "";

			string strCampi = GetTableCampi(row.Table);
			string[] strTypes = GetTableTypes(row.Table).Split(',');

			return GetRowScript(row, strCampi, strTypes);
		}

		public static string GetTableScript(DataTable tab)
		{
			if (tab == null)
				return "";

			StringBuilder strScript = new StringBuilder();

			string strCampi = GetTableCampi(tab);
			string[] strTypes = GetTableTypes(tab).Split(',');

			foreach (DataRow row in tab.Rows)
			{
				strScript.AppendLine(GetRowScript(row, strCampi, strTypes));
			}

			return strScript.ToString();
		}

		public static string GetUpdateScript(DataRow row, string strCampi, string strWhere)
		{
			string strCampiInclusi = "," + strCampi + ",";

			string strSql = "UPDATE " + row.Table.TableName + " SET ";

			string[] strCampiTab = GetTableCampi(row.Table).Split(',');
			string[] strTypes = GetTableTypes(row.Table).Split(',');

			string strValori = "";
			for (int intCol = 0; intCol < strTypes.Length; intCol++)
			{
				string strCampo = strCampiTab[intCol];
				if (!strCampiInclusi.Contains("," + strCampo + ","))
					continue;

				if (strValori != "")
					strValori += ", ";

				object objDato = row[intCol];

				strValori += strCampo + "=" + GetDatoValue(objDato, strTypes[intCol]);
			}

			strSql += strValori;
			strSql += " " + strWhere;

			return strSql;
		}

		public static void CleanRow(DataRow row)
		{
			for (int intCol = 0; intCol < row.Table.Columns.Count; intCol++)
			{
				row[intCol] = DBNull.Value;
			}
		}
	}
}
