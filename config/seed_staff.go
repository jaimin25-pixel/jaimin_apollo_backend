package config

import (
	"fmt"
	"log"
	"time"

	"apollo-backend/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RunSeedStaff is the exported entry point (used by cmd/seed and database.go).
func RunSeedStaff(db *gorm.DB) { seedStaff(db) }

// seedStaff seeds doctors, staff, and pharmacists if not already present.
func seedStaff(db *gorm.DB) {
	var docCount int64
	db.Model(&model.Doctor{}).Count(&docCount)
	if docCount > 0 {
		log.Println("staff already seeded, skipping")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("Staff@123"), 12)
	if err != nil {
		log.Printf("seed_staff: bcrypt error: %v", err)
		return
	}
	h := string(hash)

	// Build department name → ID map
	var depts []model.Department
	db.Find(&depts)
	deptID := make(map[string]uint, len(depts))
	for _, d := range depts {
		deptID[d.Name] = d.DeptID
	}

	join := func(y, m, d int) time.Time {
		return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	}

	// ── DOCTORS ──────────────────────────────────────────────────────

	type docSeed struct {
		code           string
		name           string
		email          string
		dept           string
		specialization string
		qualification  string
		phone          string
		joined         time.Time
	}

	doctors := []docSeed{
		// General Medicine (5)
		{"DOC-0001", "Dr. Rajesh Kumar", "dr.rajesh.kumar@apollo.health", "General Medicine", "General Physician", "MBBS, MD (General Medicine)", "+91-9811000001", join(2019, 3, 1)},
		{"DOC-0002", "Dr. Priya Sharma", "dr.priya.sharma@apollo.health", "General Medicine", "General Physician", "MBBS, MD", "+91-9811000002", join(2020, 6, 15)},
		{"DOC-0003", "Dr. Anil Mehta", "dr.anil.mehta@apollo.health", "General Medicine", "Internal Medicine", "MBBS, MD, DNB", "+91-9811000003", join(2018, 1, 10)},
		{"DOC-0004", "Dr. Sunita Patel", "dr.sunita.patel@apollo.health", "General Medicine", "General Medicine", "MBBS, MD", "+91-9811000004", join(2021, 9, 5)},
		{"DOC-0005", "Dr. Vikram Singh", "dr.vikram.singh@apollo.health", "General Medicine", "General Physician", "MBBS, MD, MRCP", "+91-9811000005", join(2017, 7, 20)},

		// Cardiology (5)
		{"DOC-0006", "Dr. Ramesh Iyer", "dr.ramesh.iyer@apollo.health", "Cardiology", "Interventional Cardiology", "MBBS, MD, DM (Cardiology)", "+91-9811000006", join(2016, 4, 1)},
		{"DOC-0007", "Dr. Deepika Nair", "dr.deepika.nair@apollo.health", "Cardiology", "Cardiology", "MBBS, MD, DM", "+91-9811000007", join(2020, 2, 14)},
		{"DOC-0008", "Dr. Suresh Reddy", "dr.suresh.reddy@apollo.health", "Cardiology", "Cardiac Electrophysiology", "MBBS, MD, DM, FESC", "+91-9811000008", join(2018, 11, 30)},
		{"DOC-0009", "Dr. Anita Gupta", "dr.anita.gupta@apollo.health", "Cardiology", "Non-invasive Cardiology", "MBBS, MD, DM", "+91-9811000009", join(2022, 1, 3)},
		{"DOC-0010", "Dr. Mohit Joshi", "dr.mohit.joshi@apollo.health", "Cardiology", "Cardiology", "MBBS, MD, DM", "+91-9811000010", join(2019, 8, 25)},

		// Neurology (5)
		{"DOC-0011", "Dr. Krishna Rao", "dr.krishna.rao@apollo.health", "Neurology", "Neurology", "MBBS, MD, DM (Neurology)", "+91-9811000011", join(2017, 5, 12)},
		{"DOC-0012", "Dr. Meera Nambiar", "dr.meera.nambiar@apollo.health", "Neurology", "Stroke Medicine", "MBBS, MD, DM", "+91-9811000012", join(2020, 10, 8)},
		{"DOC-0013", "Dr. Sanjay Verma", "dr.sanjay.verma@apollo.health", "Neurology", "Epileptology", "MBBS, MD, DM", "+91-9811000013", join(2018, 3, 22)},
		{"DOC-0014", "Dr. Kavitha Pillai", "dr.kavitha.pillai@apollo.health", "Neurology", "Neurology", "MBBS, MD, DM, FEAN", "+91-9811000014", join(2021, 6, 18)},
		{"DOC-0015", "Dr. Arjun Menon", "dr.arjun.menon@apollo.health", "Neurology", "Neuro-intervention", "MBBS, MD, DM", "+91-9811000015", join(2016, 12, 1)},

		// Orthopedics (5)
		{"DOC-0016", "Dr. Ravi Chopra", "dr.ravi.chopra@apollo.health", "Orthopedics", "Orthopedic Surgery", "MBBS, MS (Ortho)", "+91-9811000016", join(2018, 7, 9)},
		{"DOC-0017", "Dr. Pooja Kulkarni", "dr.pooja.kulkarni@apollo.health", "Orthopedics", "Spine Surgery", "MBBS, MS, MCh", "+91-9811000017", join(2020, 4, 23)},
		{"DOC-0018", "Dr. Amit Desai", "dr.amit.desai@apollo.health", "Orthopedics", "Joint Replacement", "MBBS, MS (Ortho), DNB", "+91-9811000018", join(2019, 9, 14)},
		{"DOC-0019", "Dr. Neha Tripathi", "dr.neha.tripathi@apollo.health", "Orthopedics", "Sports Medicine", "MBBS, MS", "+91-9811000019", join(2022, 3, 1)},
		{"DOC-0020", "Dr. Sachin Bhatt", "dr.sachin.bhatt@apollo.health", "Orthopedics", "Trauma Surgery", "MBBS, MS, DNB", "+91-9811000020", join(2017, 11, 11)},

		// Pediatrics (5)
		{"DOC-0021", "Dr. Geeta Mishra", "dr.geeta.mishra@apollo.health", "Pediatrics", "Pediatrics", "MBBS, MD (Pediatrics)", "+91-9811000021", join(2019, 2, 7)},
		{"DOC-0022", "Dr. Rahul Aggarwal", "dr.rahul.aggarwal@apollo.health", "Pediatrics", "Neonatology", "MBBS, MD, DNB", "+91-9811000022", join(2021, 5, 20)},
		{"DOC-0023", "Dr. Swati Srivastava", "dr.swati.srivastava@apollo.health", "Pediatrics", "Pediatric Cardiology", "MBBS, MD, DM", "+91-9811000023", join(2018, 8, 13)},
		{"DOC-0024", "Dr. Varun Bhatia", "dr.varun.bhatia@apollo.health", "Pediatrics", "Pediatric Neurology", "MBBS, MD, DM", "+91-9811000024", join(2020, 12, 28)},
		{"DOC-0025", "Dr. Ritu Malhotra", "dr.ritu.malhotra@apollo.health", "Pediatrics", "Pediatrics", "MBBS, MD, FIAP", "+91-9811000025", join(2016, 6, 5)},

		// Gynecology & Obstetrics (5)
		{"DOC-0026", "Dr. Shobha Krishnan", "dr.shobha.krishnan@apollo.health", "Gynecology & Obstetrics", "Obstetrics & Gynecology", "MBBS, MD (OBG)", "+91-9811000026", join(2017, 10, 17)},
		{"DOC-0027", "Dr. Lakshmi Subramaniam", "dr.lakshmi.subramaniam@apollo.health", "Gynecology & Obstetrics", "Gynecological Oncology", "MBBS, MD, DNB", "+91-9811000027", join(2020, 1, 6)},
		{"DOC-0028", "Dr. Archana Kaur", "dr.archana.kaur@apollo.health", "Gynecology & Obstetrics", "Infertility & IVF", "MBBS, MS (OBG)", "+91-9811000028", join(2019, 7, 29)},
		{"DOC-0029", "Dr. Mamta Tiwari", "dr.mamta.tiwari@apollo.health", "Gynecology & Obstetrics", "Obstetrics", "MBBS, MD", "+91-9811000029", join(2022, 4, 11)},
		{"DOC-0030", "Dr. Divya Pandey", "dr.divya.pandey@apollo.health", "Gynecology & Obstetrics", "Gynecology", "MBBS, MS, DNB", "+91-9811000030", join(2018, 5, 3)},

		// Dermatology (5)
		{"DOC-0031", "Dr. Sameera Khan", "dr.sameera.khan@apollo.health", "Dermatology", "Dermatology", "MBBS, MD (Dermatology)", "+91-9811000031", join(2020, 9, 16)},
		{"DOC-0032", "Dr. Rohit Saxena", "dr.rohit.saxena@apollo.health", "Dermatology", "Cosmetic Dermatology", "MBBS, MD", "+91-9811000032", join(2018, 4, 8)},
		{"DOC-0033", "Dr. Priti Jain", "dr.priti.jain@apollo.health", "Dermatology", "Dermatology", "MBBS, MD, DVD", "+91-9811000033", join(2021, 11, 22)},
		{"DOC-0034", "Dr. Kiran Shetty", "dr.kiran.shetty@apollo.health", "Dermatology", "Trichology", "MBBS, MD, FRCP", "+91-9811000034", join(2017, 2, 14)},
		{"DOC-0035", "Dr. Ajay Goel", "dr.ajay.goel@apollo.health", "Dermatology", "Dermato-surgery", "MBBS, MD, MCh", "+91-9811000035", join(2019, 6, 30)},

		// ENT (5)
		{"DOC-0036", "Dr. Naresh Kapoor", "dr.naresh.kapoor@apollo.health", "ENT", "ENT", "MBBS, MS (ENT)", "+91-9811000036", join(2016, 8, 19)},
		{"DOC-0037", "Dr. Seema Bhatt", "dr.seema.bhatt@apollo.health", "ENT", "Otology", "MBBS, MS (ENT), DNB", "+91-9811000037", join(2020, 3, 4)},
		{"DOC-0038", "Dr. Vinod Sharma", "dr.vinod.sharma@apollo.health", "ENT", "Head & Neck Surgery", "MBBS, MS, MCh", "+91-9811000038", join(2018, 10, 27)},
		{"DOC-0039", "Dr. Radha Menon", "dr.radha.menon@apollo.health", "ENT", "ENT", "MBBS, MS", "+91-9811000039", join(2022, 7, 2)},
		{"DOC-0040", "Dr. Tarun Malhotra", "dr.tarun.malhotra@apollo.health", "ENT", "Rhinology", "MBBS, MS (ENT)", "+91-9811000040", join(2019, 1, 15)},

		// Ophthalmology (5)
		{"DOC-0041", "Dr. Sangeeta Verma", "dr.sangeeta.verma@apollo.health", "Ophthalmology", "Ophthalmology", "MBBS, MS (Ophthalmology)", "+91-9811000041", join(2017, 9, 8)},
		{"DOC-0042", "Dr. Pankaj Agarwal", "dr.pankaj.agarwal@apollo.health", "Ophthalmology", "Retina", "MBBS, MS, DOMS", "+91-9811000042", join(2021, 4, 18)},
		{"DOC-0043", "Dr. Shreya Nair", "dr.shreya.nair@apollo.health", "Ophthalmology", "Glaucoma", "MBBS, MS, FICS", "+91-9811000043", join(2019, 12, 9)},
		{"DOC-0044", "Dr. Alok Sinha", "dr.alok.sinha@apollo.health", "Ophthalmology", "Cornea", "MBBS, MS", "+91-9811000044", join(2020, 8, 21)},
		{"DOC-0045", "Dr. Tanya Seth", "dr.tanya.seth@apollo.health", "Ophthalmology", "Pediatric Ophthalmology", "MBBS, MS, DNB", "+91-9811000045", join(2018, 2, 26)},

		// Psychiatry (5)
		{"DOC-0046", "Dr. Manish Chaudhary", "dr.manish.chaudhary@apollo.health", "Psychiatry", "Psychiatry", "MBBS, MD (Psychiatry)", "+91-9811000046", join(2016, 11, 14)},
		{"DOC-0047", "Dr. Preeti Goyal", "dr.preeti.goyal@apollo.health", "Psychiatry", "Child Psychiatry", "MBBS, MD, MRCPsych", "+91-9811000047", join(2020, 5, 7)},
		{"DOC-0048", "Dr. Sunil Bose", "dr.sunil.bose@apollo.health", "Psychiatry", "Addiction Medicine", "MBBS, MD", "+91-9811000048", join(2018, 9, 3)},
		{"DOC-0049", "Dr. Asha Menon", "dr.asha.menon@apollo.health", "Psychiatry", "Geriatric Psychiatry", "MBBS, MD, DPM", "+91-9811000049", join(2021, 2, 16)},
		{"DOC-0050", "Dr. Neeraj Agrawal", "dr.neeraj.agrawal@apollo.health", "Psychiatry", "Psychiatry", "MBBS, MD, FRANZCP", "+91-9811000050", join(2019, 7, 11)},

		// Oncology (5)
		{"DOC-0051", "Dr. Satish Choudhary", "dr.satish.choudhary@apollo.health", "Oncology", "Medical Oncology", "MBBS, MD, DM (Oncology)", "+91-9811000051", join(2017, 4, 25)},
		{"DOC-0052", "Dr. Kavya Reddy", "dr.kavya.reddy@apollo.health", "Oncology", "Surgical Oncology", "MBBS, MS, MCh", "+91-9811000052", join(2020, 11, 12)},
		{"DOC-0053", "Dr. Ashwin Pillai", "dr.ashwin.pillai@apollo.health", "Oncology", "Radiation Oncology", "MBBS, MD, DMRT", "+91-9811000053", join(2018, 6, 17)},
		{"DOC-0054", "Dr. Nalini Iyer", "dr.nalini.iyer@apollo.health", "Oncology", "Hematology-Oncology", "MBBS, MD, DM", "+91-9811000054", join(2022, 2, 8)},
		{"DOC-0055", "Dr. Pradeep Raghavan", "dr.pradeep.raghavan@apollo.health", "Oncology", "Onco-surgery", "MBBS, MS, DNB", "+91-9811000055", join(2016, 5, 30)},

		// Nephrology (5)
		{"DOC-0056", "Dr. Girish Nambiar", "dr.girish.nambiar@apollo.health", "Nephrology", "Nephrology", "MBBS, MD, DM (Nephrology)", "+91-9811000056", join(2018, 8, 6)},
		{"DOC-0057", "Dr. Usha Krishnaswamy", "dr.usha.krishnaswamy@apollo.health", "Nephrology", "Transplant Nephrology", "MBBS, MD, DM", "+91-9811000057", join(2021, 3, 14)},
		{"DOC-0058", "Dr. Hemant Deshmukh", "dr.hemant.deshmukh@apollo.health", "Nephrology", "Nephrology", "MBBS, MD, DM, FASN", "+91-9811000058", join(2019, 10, 29)},
		{"DOC-0059", "Dr. Shilpa Wagh", "dr.shilpa.wagh@apollo.health", "Nephrology", "Nephrology", "MBBS, MD, DNB", "+91-9811000059", join(2020, 7, 19)},
		{"DOC-0060", "Dr. Yusuf Khan", "dr.yusuf.khan@apollo.health", "Nephrology", "Nephrology", "MBBS, MD", "+91-9811000060", join(2017, 1, 22)},

		// Urology (5)
		{"DOC-0061", "Dr. Ramakrishna Iyengar", "dr.ramakrishna.iyengar@apollo.health", "Urology", "Urology", "MBBS, MS, MCh (Urology)", "+91-9811000061", join(2016, 9, 28)},
		{"DOC-0062", "Dr. Subhash Garg", "dr.subhash.garg@apollo.health", "Urology", "Uro-oncology", "MBBS, MS, MCh", "+91-9811000062", join(2019, 4, 10)},
		{"DOC-0063", "Dr. Deepa Saini", "dr.deepa.saini@apollo.health", "Urology", "Female Urology", "MBBS, MS, DNB", "+91-9811000063", join(2021, 8, 23)},
		{"DOC-0064", "Dr. Rajeev Sood", "dr.rajeev.sood@apollo.health", "Urology", "Andrology", "MBBS, MS, MCh", "+91-9811000064", join(2018, 12, 5)},
		{"DOC-0065", "Dr. Harish Lal", "dr.harish.lal@apollo.health", "Urology", "Pediatric Urology", "MBBS, MS (Urology)", "+91-9811000065", join(2020, 6, 3)},

		// Pulmonology (5)
		{"DOC-0066", "Dr. Venkatesh Krishna", "dr.venkatesh.krishna@apollo.health", "Pulmonology", "Pulmonology", "MBBS, MD, DM (Pulmonology)", "+91-9811000066", join(2017, 3, 18)},
		{"DOC-0067", "Dr. Leena Thomas", "dr.leena.thomas@apollo.health", "Pulmonology", "Respiratory Medicine", "MBBS, MD, FCCP", "+91-9811000067", join(2020, 9, 27)},
		{"DOC-0068", "Dr. Santosh Tiwari", "dr.santosh.tiwari@apollo.health", "Pulmonology", "Thoracic Medicine", "MBBS, MD, DM", "+91-9811000068", join(2018, 5, 14)},
		{"DOC-0069", "Dr. Rukmini Devi", "dr.rukmini.devi@apollo.health", "Pulmonology", "Pulmonology", "MBBS, MD, DNB", "+91-9811000069", join(2022, 1, 20)},
		{"DOC-0070", "Dr. Aakash Mehrotra", "dr.aakash.mehrotra@apollo.health", "Pulmonology", "Sleep Medicine", "MBBS, MD", "+91-9811000070", join(2019, 11, 7)},

		// Gastroenterology (5)
		{"DOC-0071", "Dr. Jayant Naik", "dr.jayant.naik@apollo.health", "Gastroenterology", "Gastroenterology", "MBBS, MD, DM (Gastro)", "+91-9811000071", join(2016, 7, 4)},
		{"DOC-0072", "Dr. Chetana Patil", "dr.chetana.patil@apollo.health", "Gastroenterology", "Hepatology", "MBBS, MD, DM", "+91-9811000072", join(2020, 2, 25)},
		{"DOC-0073", "Dr. Vivek Trivedi", "dr.vivek.trivedi@apollo.health", "Gastroenterology", "Endoscopy", "MBBS, MD, DNB", "+91-9811000073", join(2018, 10, 16)},
		{"DOC-0074", "Dr. Chandrika Sharma", "dr.chandrika.sharma@apollo.health", "Gastroenterology", "Gastroenterology", "MBBS, MD", "+91-9811000074", join(2021, 6, 9)},
		{"DOC-0075", "Dr. Saurabh Joshi", "dr.saurabh.joshi@apollo.health", "Gastroenterology", "Hepato-biliary", "MBBS, MD, DM, FAGE", "+91-9811000075", join(2019, 3, 2)},

		// Endocrinology (5)
		{"DOC-0076", "Dr. Krishnamurthy Rajan", "dr.krishnamurthy.rajan@apollo.health", "Endocrinology", "Endocrinology", "MBBS, MD, DM (Endocrinology)", "+91-9811000076", join(2017, 8, 21)},
		{"DOC-0077", "Dr. Padmavathi Iyer", "dr.padmavathi.iyer@apollo.health", "Endocrinology", "Diabetes & Endocrinology", "MBBS, MD, FRCP", "+91-9811000077", join(2020, 4, 13)},
		{"DOC-0078", "Dr. Gaurav Khanna", "dr.gaurav.khanna@apollo.health", "Endocrinology", "Thyroid Medicine", "MBBS, MD, DM", "+91-9811000078", join(2018, 1, 31)},
		{"DOC-0079", "Dr. Smita Rao", "dr.smita.rao@apollo.health", "Endocrinology", "Endocrinology", "MBBS, MD, DNB", "+91-9811000079", join(2022, 5, 6)},
		{"DOC-0080", "Dr. Nikhil Mathur", "dr.nikhil.mathur@apollo.health", "Endocrinology", "Metabolic Medicine", "MBBS, MD", "+91-9811000080", join(2019, 2, 19)},

		// Radiology (5)
		{"DOC-0081", "Dr. Anand Bhat", "dr.anand.bhat@apollo.health", "Radiology", "Radiology", "MBBS, MD (Radiology)", "+91-9811000081", join(2016, 10, 7)},
		{"DOC-0082", "Dr. Jyothi Kumari", "dr.jyothi.kumari@apollo.health", "Radiology", "Interventional Radiology", "MBBS, MD, FRCR", "+91-9811000082", join(2019, 6, 24)},
		{"DOC-0083", "Dr. Radhakrishnan Menon", "dr.radhakrishnan.menon@apollo.health", "Radiology", "Neuroradiology", "MBBS, MD, DNB", "+91-9811000083", join(2021, 1, 12)},
		{"DOC-0084", "Dr. Sudha Balakrishnan", "dr.sudha.balakrishnan@apollo.health", "Radiology", "Radiology", "MBBS, MD, DMRD", "+91-9811000084", join(2018, 7, 5)},
		{"DOC-0085", "Dr. Mahendra Verma", "dr.mahendra.verma@apollo.health", "Radiology", "Nuclear Medicine", "MBBS, MD", "+91-9811000085", join(2020, 11, 18)},

		// Pathology (5)
		{"DOC-0086", "Dr. Savita Dixit", "dr.savita.dixit@apollo.health", "Pathology", "Pathology", "MBBS, MD (Pathology)", "+91-9811000086", join(2017, 6, 15)},
		{"DOC-0087", "Dr. Rakesh Misra", "dr.rakesh.misra@apollo.health", "Pathology", "Hematopathology", "MBBS, MD, FIAC", "+91-9811000087", join(2019, 9, 1)},
		{"DOC-0088", "Dr. Vandana Shankar", "dr.vandana.shankar@apollo.health", "Pathology", "Clinical Pathology", "MBBS, MD, DNB", "+91-9811000088", join(2021, 4, 7)},
		{"DOC-0089", "Dr. Srinivasan Pillai", "dr.srinivasan.pillai@apollo.health", "Pathology", "Microbiology", "MBBS, MD", "+91-9811000089", join(2018, 12, 23)},
		{"DOC-0090", "Dr. Bharati Rao", "dr.bharati.rao@apollo.health", "Pathology", "Biochemistry", "MBBS, MD, DNB", "+91-9811000090", join(2020, 3, 17)},

		// Emergency Medicine (5)
		{"DOC-0091", "Dr. Krishnaswamy Gopalan", "dr.krishnaswamy.gopalan@apollo.health", "Emergency Medicine", "Emergency Medicine", "MBBS, MD (Emergency)", "+91-9811000091", join(2016, 2, 9)},
		{"DOC-0092", "Dr. Pratibha Deshpande", "dr.pratibha.deshpande@apollo.health", "Emergency Medicine", "Emergency Medicine", "MBBS, MD, FCEM", "+91-9811000092", join(2019, 8, 20)},
		{"DOC-0093", "Dr. Manoj Srivastava", "dr.manoj.srivastava@apollo.health", "Emergency Medicine", "Trauma Surgery", "MBBS, MS, ATLS", "+91-9811000093", join(2021, 5, 11)},
		{"DOC-0094", "Dr. Renuka Chakraborty", "dr.renuka.chakraborty@apollo.health", "Emergency Medicine", "Emergency Medicine", "MBBS, MD, DNB", "+91-9811000094", join(2018, 3, 28)},
		{"DOC-0095", "Dr. Aniket Kulkarni", "dr.aniket.kulkarni@apollo.health", "Emergency Medicine", "Critical Care", "MBBS, MD, EDIC", "+91-9811000095", join(2022, 6, 14)},

		// General Surgery (5)
		{"DOC-0096", "Dr. Kashinath Hegde", "dr.kashinath.hegde@apollo.health", "General Surgery", "General Surgery", "MBBS, MS (General Surgery)", "+91-9811000096", join(2017, 12, 3)},
		{"DOC-0097", "Dr. Shanthi Suresh", "dr.shanthi.suresh@apollo.health", "General Surgery", "Laparoscopic Surgery", "MBBS, MS, MCh", "+91-9811000097", join(2020, 7, 16)},
		{"DOC-0098", "Dr. Bhaskar Das", "dr.bhaskar.das@apollo.health", "General Surgery", "GI Surgery", "MBBS, MS, FAIS", "+91-9811000098", join(2018, 4, 27)},
		{"DOC-0099", "Dr. Nandini Chandra", "dr.nandini.chandra@apollo.health", "General Surgery", "Colorectal Surgery", "MBBS, MS, DNB", "+91-9811000099", join(2021, 10, 5)},
		{"DOC-0100", "Dr. Rakesh Verma", "dr.rakesh.verma@apollo.health", "General Surgery", "Vascular Surgery", "MBBS, MS, MCh", "+91-9811000100", join(2016, 3, 25)},
	}

	for _, d := range doctors {
		did, ok := deptID[d.dept]
		if !ok {
			log.Printf("seed_staff: department not found: %s", d.dept)
			continue
		}
		doc := model.Doctor{
			DocCode:        d.code,
			FullName:       d.name,
			Email:          d.email,
			HashedPassword: h,
			DeptID:         did,
			Specialization: d.specialization,
			Qualification:  d.qualification,
			Phone:          d.phone,
			JoiningDate:    d.joined,
			Status:         "active",
		}
		if err := db.Create(&doc).Error; err != nil {
			log.Printf("seed doctor %s: %v", d.email, err)
		}
	}
	log.Printf("seeded %d doctors", len(doctors))

	// ── STAFF ────────────────────────────────────────────────────────

	type staffSeed struct {
		name           string
		email          string
		role           string
		dept           string
		qualification  string
		employmentType string
		joined         time.Time
	}

	staffMembers := []staffSeed{
		// Nurses (10) — spread across clinical depts
		{"Ananya Singh", "nurse.ananya.singh@apollo.health", "nurse", "General Medicine", "B.Sc Nursing, RNRM", "full_time", join(2020, 6, 1)},
		{"Bharathi Kumar", "nurse.bharathi.kumar@apollo.health", "nurse", "Cardiology", "B.Sc Nursing, RNRM", "full_time", join(2019, 3, 15)},
		{"Chetna Sharma", "nurse.chetna.sharma@apollo.health", "nurse", "Neurology", "B.Sc Nursing, PGDCC", "full_time", join(2021, 8, 10)},
		{"Devi Pillai", "nurse.devi.pillai@apollo.health", "nurse", "Orthopedics", "GNM, RNRM", "full_time", join(2018, 11, 20)},
		{"Esha Mehta", "nurse.esha.mehta@apollo.health", "nurse", "Pediatrics", "B.Sc Nursing", "full_time", join(2022, 2, 7)},
		{"Fatima Khan", "nurse.fatima.khan@apollo.health", "nurse", "Gynecology & Obstetrics", "B.Sc Nursing, Midwifery", "full_time", join(2020, 9, 14)},
		{"Geetha Nair", "nurse.geetha.nair@apollo.health", "nurse", "Oncology", "B.Sc Nursing, Oncology Cert", "full_time", join(2017, 5, 22)},
		{"Hemalatha Iyer", "nurse.hemalatha.iyer@apollo.health", "nurse", "Emergency Medicine", "B.Sc Nursing, RNRM", "full_time", join(2019, 7, 3)},
		{"Indira Rao", "nurse.indira.rao@apollo.health", "nurse", "General Surgery", "GNM, Post-Basic BSc", "full_time", join(2021, 1, 18)},
		{"Jayalakshmi Krishna", "nurse.jayalakshmi.krishna@apollo.health", "nurse", "Nephrology", "B.Sc Nursing, Dialysis Cert", "full_time", join(2018, 4, 29)},

		// Receptionists (4) — Administration
		{"Arun Patel", "reception.arun.patel@apollo.health", "receptionist", "Administration Department", "BBA, Hospital Admin Diploma", "full_time", join(2021, 6, 7)},
		{"Beena Thomas", "reception.beena.thomas@apollo.health", "receptionist", "Administration Department", "BA, Diploma in Hospital Mgmt", "full_time", join(2019, 10, 14)},
		{"Chandra Menon", "reception.chandra.menon@apollo.health", "receptionist", "Administration Department", "B.Com", "full_time", join(2022, 3, 21)},
		{"Deepti Bajaj", "reception.deepti.bajaj@apollo.health", "receptionist", "Administration Department", "BBA", "part_time", join(2020, 8, 5)},

		// Lab Technicians (5) — Pathology
		{"Rajan Nair", "lab.rajan.nair@apollo.health", "lab_technician", "Pathology", "B.Sc MLT, DMLT", "full_time", join(2018, 7, 11)},
		{"Shanti Devi", "lab.shanti.devi@apollo.health", "lab_technician", "Pathology", "DMLT, B.Sc MLT", "full_time", join(2020, 2, 24)},
		{"Arunkumar Pillai", "lab.arunkumar.pillai@apollo.health", "lab_technician", "Pathology", "B.Sc Biochemistry, DMLT", "full_time", join(2021, 9, 17)},
		{"Bindu Krishna", "lab.bindu.krishna@apollo.health", "lab_technician", "Pathology", "B.Sc MLT", "full_time", join(2019, 4, 30)},
		{"Chandrasekhar Rao", "lab.chandrasekhar.rao@apollo.health", "lab_technician", "Pathology", "B.Sc Microbiology, DMLT", "full_time", join(2017, 12, 8)},

		// Radiologists (3) — Radiology
		{"Karthik Subramaniam", "radio.karthik.subramaniam@apollo.health", "radiologist", "Radiology", "B.Sc Radiography, PGDRT", "full_time", join(2019, 5, 6)},
		{"Malini Iyer", "radio.malini.iyer@apollo.health", "radiologist", "Radiology", "B.Sc Radiography", "full_time", join(2021, 11, 15)},
		{"Naveen Sharma", "radio.naveen.sharma@apollo.health", "radiologist", "Radiology", "B.Sc Radiography, CT/MRI Cert", "full_time", join(2018, 3, 26)},

		// Billing Staff (3) — Administration
		{"Anjali Khanna", "billing.anjali.khanna@apollo.health", "billing_staff", "Administration Department", "B.Com, Medical Billing Cert", "full_time", join(2020, 10, 12)},
		{"Bhupesh Gupta", "billing.bhupesh.gupta@apollo.health", "billing_staff", "Administration Department", "B.Com, CA (Inter)", "full_time", join(2018, 6, 19)},
		{"Chetan Agarwal", "billing.chetan.agarwal@apollo.health", "billing_staff", "Administration Department", "MBA (Finance)", "full_time", join(2022, 4, 4)},

		// HR Managers (2) — Administration
		{"Devika Sharma", "hr.devika.sharma@apollo.health", "hr_manager", "Administration Department", "MBA (HR)", "full_time", join(2017, 9, 23)},
		{"Eshwar Rao", "hr.eshwar.rao@apollo.health", "hr_manager", "Administration Department", "PGDM (HR), MSW", "full_time", join(2020, 1, 28)},

		// IT Admins (2) — Administration
		{"Farouk Sheikh", "it.farouk.sheikh@apollo.health", "it_admin", "Administration Department", "B.Tech (CS), CCNA", "full_time", join(2019, 7, 9)},
		{"Gayatri Pillai", "it.gayatri.pillai@apollo.health", "it_admin", "Administration Department", "B.Tech (IT), AWS Certified", "full_time", join(2021, 3, 16)},

		// Ward Boys (6) — various clinical depts
		{"Harish Nair", "ward.harish.nair@apollo.health", "ward_boy", "General Medicine", "10+2", "full_time", join(2020, 11, 2)},
		{"Ibrahim Khan", "ward.ibrahim.khan@apollo.health", "ward_boy", "Cardiology", "10+2, First Aid Cert", "full_time", join(2019, 6, 27)},
		{"Jagdish Kumar", "ward.jagdish.kumar@apollo.health", "ward_boy", "Orthopedics", "10+2", "full_time", join(2022, 8, 13)},
		{"Krishna Murthy", "ward.krishna.murthy@apollo.health", "ward_boy", "General Surgery", "10+2", "full_time", join(2018, 2, 17)},
		{"Laxman Yadav", "ward.laxman.yadav@apollo.health", "ward_boy", "Emergency Medicine", "10+2, BLS Cert", "full_time", join(2021, 7, 21)},
		{"Mukesh Prajapati", "ward.mukesh.prajapati@apollo.health", "ward_boy", "Pediatrics", "10+2", "full_time", join(2020, 4, 8)},

		// Housekeeping (4) — Support Services
		{"Nalini Rao", "house.nalini.rao@apollo.health", "housekeeping", "Support Services Department", "10+2", "full_time", join(2019, 12, 3)},
		{"Omprakash Verma", "house.omprakash.verma@apollo.health", "housekeeping", "Support Services Department", "10+2", "full_time", join(2021, 5, 30)},
		{"Padma Krishnan", "house.padma.krishnan@apollo.health", "housekeeping", "Support Services Department", "10th Grade", "part_time", join(2022, 9, 19)},
		{"Ramadevi Naidu", "house.ramadevi.naidu@apollo.health", "housekeeping", "Support Services Department", "10+2", "full_time", join(2018, 8, 14)},

		// Security (4) — Support Services
		{"Suresh Babu", "security.suresh.babu@apollo.health", "security", "Support Services Department", "12th Grade, Security Training", "full_time", join(2020, 3, 11)},
		{"Thomas Joseph", "security.thomas.joseph@apollo.health", "security", "Support Services Department", "12th Grade, Ex-Army", "full_time", join(2017, 11, 26)},
		{"Umesh Mishra", "security.umesh.mishra@apollo.health", "security", "Support Services Department", "12th Grade, Security Cert", "full_time", join(2021, 2, 4)},
		{"Venkatesan Kumar", "security.venkatesan.kumar@apollo.health", "security", "Support Services Department", "B.A, Security Training", "full_time", join(2019, 8, 18)},

		// Paramedics (4) — Emergency
		{"Waqar Ahmed", "para.waqar.ahmed@apollo.health", "paramedic", "Emergency Medicine", "Diploma in Paramedic, EMT-B", "full_time", join(2020, 5, 15)},
		{"Xavier Fernandez", "para.xavier.fernandez@apollo.health", "paramedic", "Emergency Medicine", "B.Sc Paramedics, ACLS", "full_time", join(2018, 10, 22)},
		{"Yogesh Sharma", "para.yogesh.sharma@apollo.health", "paramedic", "Emergency Medicine", "Diploma in Paramedic", "full_time", join(2022, 7, 7)},
		{"Zainab Shaikh", "para.zainab.shaikh@apollo.health", "paramedic", "Emergency Medicine", "B.Sc Paramedics, PHTLS", "full_time", join(2019, 1, 9)},
	}

	for _, s := range staffMembers {
		var did *uint
		if id, ok := deptID[s.dept]; ok {
			did = &id
		} else {
			log.Printf("seed_staff: dept not found for staff %s: %s", s.email, s.dept)
		}
		member := model.Staff{
			FullName:       s.name,
			Email:          s.email,
			HashedPassword: h,
			Role:           s.role,
			DeptID:         did,
			Qualification:  s.qualification,
			EmploymentType: s.employmentType,
			JoiningDate:    s.joined,
			Status:         "active",
		}
		if err := db.Create(&member).Error; err != nil {
			log.Printf("seed staff %s: %v", s.email, err)
		}
	}
	log.Printf("seeded %d staff members", len(staffMembers))

	// ── PHARMACISTS ──────────────────────────────────────────────────

	type pharmSeed struct {
		name    string
		email   string
		license string
		phone   string
	}

	pharmacists := []pharmSeed{
		{"Aditya Raj", "pharma.aditya.raj@apollo.health", "PH-2024-001", "+91-9899000001"},
		{"Bhavna Mehta", "pharma.bhavna.mehta@apollo.health", "PH-2024-002", "+91-9899000002"},
		{"Chandramouli Iyer", "pharma.chandramouli.iyer@apollo.health", "PH-2024-003", "+91-9899000003"},
		{"Daksha Shah", "pharma.daksha.shah@apollo.health", "PH-2024-004", "+91-9899000004"},
		{"Eknath Rao", "pharma.eknath.rao@apollo.health", "PH-2024-005", "+91-9899000005"},
		{"Faheema Begum", "pharma.faheema.begum@apollo.health", "PH-2024-006", "+91-9899000006"},
	}

	for _, p := range pharmacists {
		pharm := model.Pharmacist{
			FullName:       p.name,
			Email:          p.email,
			HashedPassword: h,
			LicenseNumber:  p.license,
			Phone:          p.phone,
			Status:         "active",
		}
		if err := db.Create(&pharm).Error; err != nil {
			log.Printf("seed pharmacist %s: %v", p.email, err)
		}
	}
	log.Printf("seeded %d pharmacists", len(pharmacists))

	// Update HOD doctors for departments (first doctor in each dept becomes HOD)
	seedHODs(db)
}

// seedHODs sets the Head of Department for each clinical department.
func seedHODs(db *gorm.DB) {
	hodMap := map[string]string{
		"General Medicine":        "dr.rajesh.kumar@apollo.health",
		"Cardiology":              "dr.ramesh.iyer@apollo.health",
		"Neurology":               "dr.krishna.rao@apollo.health",
		"Orthopedics":             "dr.ravi.chopra@apollo.health",
		"Pediatrics":              "dr.geeta.mishra@apollo.health",
		"Gynecology & Obstetrics": "dr.shobha.krishnan@apollo.health",
		"Dermatology":             "dr.sameera.khan@apollo.health",
		"ENT":                     "dr.naresh.kapoor@apollo.health",
		"Ophthalmology":           "dr.sangeeta.verma@apollo.health",
		"Psychiatry":              "dr.manish.chaudhary@apollo.health",
		"Oncology":                "dr.satish.choudhary@apollo.health",
		"Nephrology":              "dr.girish.nambiar@apollo.health",
		"Urology":                 "dr.ramakrishna.iyengar@apollo.health",
		"Pulmonology":             "dr.venkatesh.krishna@apollo.health",
		"Gastroenterology":        "dr.jayant.naik@apollo.health",
		"Endocrinology":           "dr.krishnamurthy.rajan@apollo.health",
		"Radiology":               "dr.anand.bhat@apollo.health",
		"Pathology":               "dr.savita.dixit@apollo.health",
		"Emergency Medicine":      "dr.krishnaswamy.gopalan@apollo.health",
		"General Surgery":         "dr.kashinath.hegde@apollo.health",
	}

	for deptName, hodEmail := range hodMap {
		var doc model.Doctor
		if err := db.Where("email = ?", hodEmail).First(&doc).Error; err != nil {
			continue
		}
		db.Model(&model.Department{}).
			Where("name = ?", deptName).
			Update("hod_doctor_id", doc.DoctorID)
	}
	log.Printf("assigned HODs for %d departments", len(hodMap))
}

// padNumber returns n as a zero-padded string of width digits.
func padNumber(n, width int) string {
	return fmt.Sprintf("%0*d", width, n)
}
