using Supporto;
using System.Reflection.Emit;
using System.Reflection;
using System.Text;
using System.Xml.Serialization;
using DbManager;
using DbManager.Utility;
using Microsoft.Extensions.Logging;
using System.Xml.Schema;
using System.Xml;
using Datex2Reference;
using NLog;
using NLog.Web;

namespace DatexModelOrig
{
	public class TraduttoreSituation
	{
		#region DATA
		static Logger logger = NLogBuilder.ConfigureNLog("NLog.config").GetLogger("TraduttoreSituation");
		postgresContext db;
		public string PublicationCreator { get; set; } = "";
		#endregion

		public TraduttoreSituation(postgresContext db)
		{
			this.db = db;
		}

		public string GetD2LogicalModelXml(List<Datex2Event> eventi, int intIdDestinatario)
		{
			try
			{
				logger.Trace("in");

				D2LogicalModel? objModel = GetD2LogicalModel(eventi, intIdDestinatario);

				// serializzo tutta la pubblicazione
				if (objModel != null)
				{
					XmlSerializer SerializerObj = new XmlSerializer(typeof(D2LogicalModel));
					MemoryStream objStream = new MemoryStream();
					SerializerObj.Serialize(objStream, objModel);
					return Encoding.UTF8.GetString(objStream.ToArray());
				}

				return "";
			}
			catch (Exception ex)
			{
				logger.Error(ex.Message);
				return "";
			}
			finally
			{
				logger.Trace("out");
			}
		}

		public D2LogicalModel? GetD2LogicalModel(List<Datex2Event> eventi, int intIdDestinatario)
		{
			try
			{
				logger.Trace("in");

				// mi accerto che gli eventi siano ordinati per situazione ed evento, in modo che raggruppi gli eventi nella stessa situazione
				eventi = eventi.OrderBy(x => x.Situazione).ThenBy(x => x.Id).ToList();

				// se nell'XML ci sono tag non previsti nella classe generata dall'XSD del comparto italiano, si generano delle eccezioni, oppure non vengono valorizzati correttamente tutti i dati nella classe
				D2LogicalModel objModel = new D2LogicalModel();
				{
					// exchange
					objModel.exchange = new Exchange();
					objModel.exchange.supplierIdentification = new InternationalIdentifier();
					objModel.exchange.supplierIdentification.country = CountryEnum.it;
					objModel.exchange.supplierIdentification.nationalIdentifier = PublicationCreator;

					// situation pubblication
					// lista situazioni
					List<Situation> lstSituation = new List<Situation>();
					SituationPublication objSituationPubblication = new SituationPublication();
					{
						// payload pubblication
						objModel.payloadPublication = objSituationPubblication;
						objModel.payloadPublication.lang = "it";
						objModel.payloadPublication.publicationTime = DateTime.Now;
						objModel.payloadPublication.publicationCreator = new InternationalIdentifier();
						objModel.payloadPublication.publicationCreator.country = CountryEnum.it;
						objModel.payloadPublication.publicationCreator.nationalIdentifier = PublicationCreator;

						// situation records
						List<SituationRecord> lstSituationRecord = new List<SituationRecord>();
						Situation? objSituation = null;
						foreach (Datex2Event evento in eventi)
						{
							logger.Debug("Evento " + evento.Id);
							SituationRecord? record = null;

							bool sottotipoFound = false;
							TAB_SOTTOTIPI? sottotipo = GetTraduciSottotipo("GENERIC", evento.Categoria);
							if (sottotipo != null && sottotipo.Abilitato)
							{
								if (sottotipo.Classname == "SpeedManagement") record = GetSpeedManagement(sottotipo);
								else if (sottotipo.Classname == "MaintenanceWorks") record = GetMaintenanceWorks(sottotipo);
								else if (sottotipo.Classname == "AbnormalTraffic") record = GetAbnormalTraffic(sottotipo);
								else if (sottotipo.Classname == "RoadOrCarriagewayOrLaneManagement") record = GetRoadOrCarriagewayOrLaneManagement(sottotipo);
								else if (sottotipo.Classname == "GeneralObstruction") record = GetGeneralObstruction(sottotipo);
								else if (sottotipo.Classname == "WinterDrivingManagement") record = GetWinterDrivingManagement(sottotipo);
								else if (sottotipo.Classname == "PoorEnvironmentConditions") record = GetPoorEnvironmentConditions(sottotipo);
								else if (sottotipo.Classname == "Accident") record = GetAccident(sottotipo);
								else if (sottotipo.Classname == "PublicEvent") record = GetPublicEvent(sottotipo);

								if (record != null)
								{
									sottotipoFound = true;
									SetSituationRecord(record, evento);
								}
							}

							if (!sottotipoFound)
							{
								// sottotipo non tradotto
								if (sottotipo == null || (sottotipo != null && sottotipo.Abilitato))
								{
									evento.elab.erroreFormattazione = "Sottotipo " + evento.Categoria + " non tradotto";
									logger.Warn(evento.elab.erroreFormattazione);
								}
							}
							else
							{
								// situation
								if (objSituation == null || objSituation.id != evento.Situazione)
								{
									if (objSituation != null)
									{
										foreach (SituationRecord rec in lstSituationRecord)
											objSituation.situationRecord.Add(rec);

										lstSituationRecord.Clear();
									}

									objSituation = new Situation();
									//objSituation.situationRecord = null;
									objSituation.id = evento.Situazione;
									objSituation.version = "1"; //$DIA: evento.intVersioneSituazione.ToString();
									objSituation.headerInformation = new HeaderInformation();
									objSituation.headerInformation.confidentiality = ConfidentialityValueEnum.noRestriction;
									objSituation.headerInformation.informationStatus = InformationStatusEnum.real;
									objSituation.headerInformation.urgency = UrgencyEnum.normalUrgency;

									objSituation.overallSeveritySpecified = true;
									if (sottotipo != null && sottotipo.Severity != null)
										objSituation.overallSeverity = ParseEnum<SeverityEnum>(sottotipo.Severity);
									else
										objSituation.overallSeverity = SeverityEnum.unknown;

									lstSituation.Add(objSituation);
								}

								// aggiungo il situation record alla situation corrente
								if (record != null)
									lstSituationRecord.Add(record);
							}
						}

						if (objSituation == null)
							objSituation = new Situation();

						foreach (SituationRecord rec in lstSituationRecord)
						{
							objSituation.situationRecord.Add(rec);
						}
					}

					foreach (Situation sit in lstSituation)
					{
						objSituationPubblication.situation.Add(sit);
					}
				}

				return objModel;
			}
			catch (Exception ex)
			{
				logger.Error(ex.Message);
				return null;
			}
			finally
			{
				logger.Trace("out");
			}
		}

		private TAB_SOTTOTIPI? GetTraduciSottotipo(string tipo, string sottotipo)
		{
			try
			{
				List<TAB_SOTTOTIPI> sottotipi = DbUtility.Sottotipi_LoadAll(db);

				TAB_SOTTOTIPI? trad = sottotipi.FirstOrDefault(x => x.TypeCode == "GENERIC" && x.SubtypeCode == sottotipo && x.Uscita);
				return trad;
			}
			catch (Exception ex)
			{
				logger.Error(ex.Message);
				return null;
			}
			finally
			{
				logger.Trace("out");
			}
		}

		private SituationRecord? GetSpeedManagement(TAB_SOTTOTIPI sottotipo)
		{
			try
			{
				logger.Trace("in");

				SpeedManagement record = new SpeedManagement();
				record.complianceOption = ComplianceOptionEnum.mandatory;

				if (!string.IsNullOrEmpty(sottotipo.TypeValue))
				{
					record.speedManagementType = ParseEnum<SpeedManagementTypeEnum>(sottotipo.TypeValue);
				}

				return record;
			}
			catch (Exception ex)
			{
				logger.Error(ex.Message);
				return null;
			}
			finally
			{
				logger.Trace("out");
			}
		}

		private SituationRecord? GetMaintenanceWorks(TAB_SOTTOTIPI sottotipo)
		{
			try
			{
				logger.Trace("in");

				MaintenanceWorks record = new MaintenanceWorks();

				if (!string.IsNullOrEmpty(sottotipo.TypeValue))
				{
					record.roadMaintenanceType.Add(ParseEnum<RoadMaintenanceTypeEnum>(sottotipo.TypeValue));
				}

				return record;
			}
			catch (Exception ex)
			{
				logger.Error(ex.Message);
				return null;
			}
			finally
			{
				logger.Trace("out");
			}
		}

		private SituationRecord? GetAbnormalTraffic(TAB_SOTTOTIPI sottotipo)
		{
			try
			{
				logger.Trace("in");

				AbnormalTraffic record = new AbnormalTraffic();

				if (!string.IsNullOrEmpty(sottotipo.TypeValue))
				{
					record.abnormalTrafficType = ParseEnum<AbnormalTrafficTypeEnum>(sottotipo.TypeValue);
					record.abnormalTrafficTypeSpecified = true;
				}

				return record;
			}
			catch (Exception ex)
			{
				logger.Error(ex.Message);
				return null;
			}
			finally
			{
				logger.Trace("out");
			}
		}

		private SituationRecord? GetRoadOrCarriagewayOrLaneManagement(TAB_SOTTOTIPI sottotipo)
		{
			try
			{
				logger.Trace("in");

				RoadOrCarriagewayOrLaneManagement record = new RoadOrCarriagewayOrLaneManagement();
				record.complianceOption = ComplianceOptionEnum.mandatory;

				if (!string.IsNullOrEmpty(sottotipo.TypeValue))
				{
					record.roadOrCarriagewayOrLaneManagementType = ParseEnum<RoadOrCarriagewayOrLaneManagementTypeEnum>(sottotipo.TypeValue);
				}

				return record;
			}
			catch (Exception ex)
			{
				logger.Error(ex.Message);
				return null;
			}
			finally
			{
				logger.Trace("out");
			}
		}

		private SituationRecord? GetGeneralObstruction(TAB_SOTTOTIPI sottotipo)
		{
			try
			{
				logger.Trace("in");

				GeneralObstruction record = new GeneralObstruction();

				if (!string.IsNullOrEmpty(sottotipo.TypeValue))
				{
					record.obstructionType.Add(ParseEnum<ObstructionTypeEnum>(sottotipo.TypeValue));
				}

				return record;
			}
			catch (Exception ex)
			{
				logger.Error(ex.Message);
				return null;
			}
			finally
			{
				logger.Trace("out");
			}
		}

		private SituationRecord? GetWinterDrivingManagement(TAB_SOTTOTIPI sottotipo)
		{
			try
			{
				logger.Trace("in");

				WinterDrivingManagement record = new WinterDrivingManagement();

				if (!string.IsNullOrEmpty(sottotipo.TypeValue))
				{
					record.winterEquipmentManagementType = ParseEnum<WinterEquipmentManagementTypeEnum>(sottotipo.TypeValue);

					if (sottotipo.ExtraAttribute != null && sottotipo.ExtraValue != null)
					{
						if (sottotipo.ExtraAttribute == "complianceOption")
							record.complianceOption = ParseEnum<ComplianceOptionEnum>(sottotipo.ExtraValue);
					}
				}

				return record;
			}
			catch (Exception ex)
			{
				logger.Error(ex.Message);
				return null;
			}
			finally
			{
				logger.Trace("out");
			}
		}

		private SituationRecord? GetPoorEnvironmentConditions(TAB_SOTTOTIPI sottotipo)
		{
			try
			{
				logger.Trace("in");

				PoorEnvironmentConditions record = new PoorEnvironmentConditions();

				if (!string.IsNullOrEmpty(sottotipo.TypeValue))
				{
					record.poorEnvironmentType.Add(ParseEnum<PoorEnvironmentTypeEnum>(sottotipo.TypeValue));
				}

				return record;
			}
			catch (Exception ex)
			{
				logger.Error(ex.Message);
				return null;
			}
			finally
			{
				logger.Trace("out");
			}
		}

		private SituationRecord? GetAccident(TAB_SOTTOTIPI sottotipo)
		{
			try
			{
				logger.Trace("in");

				Accident record = new Accident();

				if (!string.IsNullOrEmpty(sottotipo.TypeValue))
				{
					record.accidentType.Add(ParseEnum<AccidentTypeEnum>(sottotipo.TypeValue));
				}

				return record;
			}
			catch (Exception ex)
			{
				logger.Error(ex.Message);
				return null;
			}
			finally
			{
				logger.Trace("out");
			}
		}

		private SituationRecord? GetPublicEvent(TAB_SOTTOTIPI sottotipo)
		{
			try
			{
				logger.Trace("in");

				PublicEvent record = new PublicEvent();

				if (!string.IsNullOrEmpty(sottotipo.TypeValue))
				{
					record.publicEventType = ParseEnum<PublicEventTypeEnum>(sottotipo.TypeValue);
				}

				return record;
			}
			catch (Exception ex)
			{
				logger.Error(ex.Message);
				return null;
			}
			finally
			{
				logger.Trace("out");
			}
		}


		// DATI PRINCIPALI EVENTO
		private void SetSituationRecord(SituationRecord? record, Datex2Event evento)
		{
			try
			{
				logger.Trace("in");

				if (record != null)
				{
					record.id = evento.Id;
					record.situationRecordCreationReference = evento.Reference;
					record.version = evento.Version.ToString();

					SetSituationRecord_Time(record, evento);
					SetSituationRecord_Source(record, evento);
					SetSituationRecord_ValidityTimeSpec(record, evento);
					SetSituationRecord_Comment(record, evento);
					SetSituationRecord_GroupOfLocation(record, evento);
				}
			}
			catch (Exception ex)
			{
				evento.elab.erroreFormattazione = "Errore: " + ex.Message;
				logger.Error(ex.Message);
			}
			finally
			{
				logger.Trace("out");
			}
		}

		private void SetSituationRecord_Time(SituationRecord record, Datex2Event evento)
		{
			try
			{
				logger.Trace("in");

				record.situationRecordCreationTime = Convert.ToDateTime(evento.CreationTime);
				record.situationRecordVersionTime = Convert.ToDateTime(evento.VersionTime).ToLocalTime();
			}
			catch (Exception ex)
			{
				evento.elab.erroreFormattazione = "Errore: " + ex.Message;
				logger.Error(ex.Message);
			}
			finally
			{
				logger.Trace("out");
			}
		}

		private void SetSituationRecord_Source(SituationRecord record, Datex2Event evento)
		{
			try
			{
				logger.Trace("in");

				// in uscita mettere sempre roadAuthority
				string strFonteType = "roadAuthorities";
				if (strFonteType != "")
				{
					record.source = new Source();
					record.source.sourceType = ParseEnum<SourceTypeEnum>(strFonteType);
					record.source.sourceTypeSpecified = true;
					record.source.reliable = true;
					record.source.reliableSpecified = true;
					record.source.sourceCountry = CountryEnum.it;
					record.source.sourceCountrySpecified = true;
					record.source.sourceIdentification = PublicationCreator;

					record.source.sourceName = new MultilingualString();
					record.source.sourceName.values.Add(GetMultilingualStringValue("it", evento.SourceName));
				}
			}
			catch (Exception ex)
			{
				evento.elab.erroreFormattazione = "Errore: " + ex.Message;
				logger.Error(ex.Message);
			}
			finally
			{
				logger.Trace("out");
			}
		}

		private void SetSituationRecord_ValidityTimeSpec(SituationRecord record, Datex2Event evento)
		{
			try
			{
				logger.Trace("in");

				record.validity = new Validity();

				// stato evento
				if (evento.StartTime <= DateTime.Now && evento.EndTime > DateTime.Now)
				{
					record.validity.validityStatus = ValidityStatusEnum.active;
				}
				else if (evento.StartTime > DateTime.Now)
				{
					record.validity.validityStatus = ValidityStatusEnum.definedByValidityTimeSpec;
					record.probabilityOfOccurrence = ProbabilityOfOccurrenceEnum.probable;
				}
				else if (evento.EndTime <=  DateTime.Now)
				{
					record.validity.validityStatus = ValidityStatusEnum.definedByValidityTimeSpec;
					record.validity.overrunning = true;
					record.validity.overrunningSpecified = true;
				}

				// data inizio evento
				if (evento.StartTime != DateTime.MinValue)
				{
					record.validity.validityTimeSpecification = new OverallPeriod();
					record.validity.validityTimeSpecification.overallStartTime = evento.StartTime;
				}

				// data fine evento
				if (evento.EndTime != null && evento.EndTime != DateTime.MinValue && evento.EndTime != DateTime.MaxValue)
				{
					record.validity.validityTimeSpecification.overallEndTime = evento.EndTime.Value;
					record.validity.validityTimeSpecification.overallEndTimeSpecified = true;
				}

				//// fascia oraria //$DIA
				//if (evento.objDettaglio.FasciaOraria == true)
				//{
				//	TimePeriodByHour objTimePeriodByHour = new TimePeriodByHour();
				//	objTimePeriodByHour.startTimeOfPeriod = evento.objDettaglio.DataInizio;
				//	objTimePeriodByHour.endTimeOfPeriod = Utility.Conv(evento.objDettaglio.DataFine);
				//	Period objPeriod = new Period();
				//	objPeriod.recurringTimePeriodOfDay = new TimePeriodOfDay[1];
				//	objPeriod.recurringTimePeriodOfDay[0] = objTimePeriodByHour;
				//	record.validity.validityTimeSpecification.validPeriod = new Period[1];
				//	record.validity.validityTimeSpecification.validPeriod[0] = objPeriod;
				//}
			}
			catch (Exception ex)
			{
				evento.elab.erroreFormattazione = "Errore: " + ex.Message;
				logger.Error(ex.Message);
			}
			finally
			{
				logger.Trace("out");
			}
		}

		private void SetSituationRecord_Comment(SituationRecord record, Datex2Event evento)
		{
			try
			{
				logger.Trace("in");

				AddComment(record, evento.GeneralPublicCommentIt, evento.GeneralPublicCommentDe, CommentTypeEnum.description);
			}
			catch (Exception ex)
			{
				evento.elab.erroreFormattazione = "Errore: " + ex.Message;
				logger.Error(ex.Message);
			}
			finally
			{
				logger.Trace("out");
			}
		}

		private void SetSituationRecord_GroupOfLocation(SituationRecord record, Datex2Event evento)
		{
			try
			{
				logger.Trace("in");

				Point point = new Point();
				point.pointByCoordinates = new PointByCoordinates();
				point.pointByCoordinates.pointCoordinates = new PointCoordinates();
				point.pointByCoordinates.pointCoordinates.latitude = Convert.ToSingle(evento.Latitude);
				point.pointByCoordinates.pointCoordinates.longitude = Convert.ToSingle(evento.Longitude);

				record.groupOfLocations = point;
			}
			catch (Exception ex)
			{
				evento.elab.erroreFormattazione = MethodBase.GetCurrentMethod() + "> Errore: " + ex.Message;
				logger.Error(ex.Message);
			}
			finally
			{
				logger.Trace("out");
			}
		}


		// UTILITY
		private bool AddComment(SituationRecord record, string testoIt, string testoDe, CommentTypeEnum type)
		{
			try
			{
				logger.Trace("in");

				Comment obj = new Comment();
				obj.comment = new MultilingualString();
				obj.commentDateTime = record.situationRecordVersionTime;
				obj.commentDateTimeSpecified = true;
				obj.commentType = type;
				obj.commentTypeSpecified = true;
				obj.comment.values.Add(GetMultilingualStringValue("it", testoIt));
				obj.comment.values.Add(GetMultilingualStringValue("de", testoDe));

				record.generalPublicComment.Add(obj);
				return true;
			}
			catch (Exception ex)
			{
				logger.Error(ex.Message);
				return false;
			}
			finally
			{
				logger.Trace("out");
			}
		}

		private MultilingualStringValue? GetMultilingualStringValue(string strLang, string strValue)
		{
			try
			{
				logger.Trace("in");

				MultilingualStringValue objMultilingualStringValue = new MultilingualStringValue();
				objMultilingualStringValue.lang = strLang;
				objMultilingualStringValue.Value = strValue;
				return objMultilingualStringValue;
			}
			catch (Exception ex)
			{
				logger.Error(ex.Message);
				return null;
			}
			finally
			{
				logger.Trace("out");
			}
		}

		public static T ParseEnum<T>(string value)
		{
			return (T)Enum.Parse(typeof(T), value, true);
		}

		public bool ValidateXmlForXsd(string xmlPath, string xsdPath)
		{
			bool isValid = false;

			XmlDocument xml = new XmlDocument();
			xml.Load(xmlPath);
			xml.Schemas.Add(null, xsdPath);

			try
			{
				xml.Validate(null);
				isValid = true;
			}
			catch (XmlSchemaValidationException ex)
			{
				logger.Error(ex.Message);
				isValid = false;
			}

			return isValid;
		}
	}
}
