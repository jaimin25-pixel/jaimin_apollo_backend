package model

import "time"

// RadiologyOrder tracks imaging/radiology orders.
// Modalities: xray, ultrasound, ct_scan, mri, echocardiogram.
type RadiologyOrder struct {
	RadiologyID        uint       `gorm:"primaryKey;autoIncrement" json:"radiology_id"`
	PatientID          uint       `gorm:"not null;index" json:"patient_id"`
	DoctorID           uint       `gorm:"not null;index" json:"doctor_id"`
	Modality           string     `gorm:"not null;size:50" json:"modality"`
	BodyPart           string     `gorm:"size:100" json:"body_part,omitempty"`
	ClinicalIndication string     `gorm:"type:text" json:"clinical_indication,omitempty"`
	OrderedAt          time.Time  `gorm:"not null;default:now()" json:"ordered_at"`
	RadiologistID      *uint      `gorm:"index" json:"radiologist_id,omitempty"`
	ReportFilePath     string     `gorm:"type:text" json:"report_file_path,omitempty"`
	ImageFilePath      string     `gorm:"type:text" json:"image_file_path,omitempty"`
	ReportText         string     `gorm:"type:text" json:"report_text,omitempty"`
	ReportedAt         *time.Time `json:"reported_at,omitempty"`
	Status             string     `gorm:"not null;default:'ordered';size:20" json:"status"`

	// Relations
	Patient     Patient `gorm:"foreignKey:PatientID;references:PatientID" json:"patient,omitempty"`
	Doctor      Doctor  `gorm:"foreignKey:DoctorID;references:DoctorID" json:"doctor,omitempty"`
	Radiologist *Staff  `gorm:"foreignKey:RadiologistID;references:StaffID" json:"radiologist,omitempty"`
}

func (RadiologyOrder) TableName() string { return "radiology_orders" }
