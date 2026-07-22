using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Supporto
{
	public class Datex2Event
	{
		public string Id { get; set; } = "";
		public string Situazione { get; set; } = "";
		public string Reference { get; set; } = "";
		public int Version { get; set; } = 0;
		public string Categoria { get; set; } = "";
		public DateTime CreationTime { get; set; } = DateTime.MinValue;
		public DateTime VersionTime { get; set; } = DateTime.MinValue;
		public DateTime StartTime { get; set; } = DateTime.MinValue;
		public DateTime? EndTime { get; set; } = null;
		public double Latitude { get; set; } = 0.0;
		public double Longitude { get; set; } = 0.0;
		public string GeneralPublicCommentIt { get; set; } = "";
		public string GeneralPublicCommentEn { get; set; } = "";
		public string GeneralPublicCommentDe { get; set; } = "";
		public string GeneralPublicCommentType { get; set; } = "";
		public string OverallSeverity { get; set; } = "";
		public string AreaOfInterest { get; set; } = "";
		public string Confidentiality { get; set; } = "";
		public string InformationStatus { get; set; } = "";
		public string Urgency { get; set; } = "";
		public string Probability { get; set; } = "";
		public string SourceCountry { get; set; } = "";
		public string SourceName { get; set; } = "";
		public string SourceType { get; set; } = "";
		public bool SourceReliable { get; set; } = false;

		// dati aggiunti per elaborazione evento
		public EventoElab elab { get; set; } = new EventoElab();
	}

	public class EventoElab
	{
		public string erroreFormattazione { get; set; } = "";
	}
}
