package itswizard_m_resync

import (
	"github.com/itslearninggermany/itswizard_m_basic"
	"github.com/itslearninggermany/itswizard_m_imses"
	"github.com/itslearninggermany/itswizard_m_syncCache"
	"github.com/jinzhu/gorm"
	"log"
	"os"
)

type Resync struct {
	username       string
	password       string
	endpoint       string
	organisationID uint
	institutionID  uint
	persons        []itswizard_m_basic.Person
	groups         []itswizard_m_basic.Group
	memberships    []itswizard_m_basic.Membership
	DB             *gorm.DB
}

type ResyncInput struct {
	Username       string
	Password       string
	Endpoint       string
	OrganisationID uint
	IntstitutionID uint
	DB             *gorm.DB
}

//Bei großen Installationen die Gruppen mitgeben

func NewResync(input ResyncInput) *Resync {
	resync := new(Resync)
	resync.username = input.Username
	resync.password = input.Password
	resync.endpoint = input.Endpoint
	resync.organisationID = input.OrganisationID
	resync.institutionID = input.IntstitutionID
	resync.DB = input.DB
	return resync
}

/*
One whole site
*/
func (p *Resync) SettDatafromInstitution(numberOfUsers int) *Resync {

	itsl := itswizard_m_imses.NewImsesService(itswizard_m_imses.NewImsesServiceInput{
		Username: p.username,
		Password: p.password,
		Url:      p.endpoint,
	})
	output := itsl.ReadAllPersons(itswizard_m_imses.ReadAllPersonsInput{
		PageIndex:                1,
		PageSize:                 numberOfUsers,
		CreatedFrom:              "",
		OnlyManuallyCreatedUsers: false,
		ConvertFromManual:        false,
	})
	if output.Err != nil {
		log.Println(output.Err)
		os.Exit(123)
	}
	alleGruppenIds := make(map[string]bool)
	for _, person := range output.Persons {
		p.persons = append(p.persons, itswizard_m_basic.Person{
			PersonSyncKey:  person.SyncPersonKey,
			FirstName:      person.FirstName,
			LastName:       person.LastName,
			Username:       person.Username,
			Profile:        person.Profile,
			Email:          person.Email,
			Phone:          person.Phone,
			Mobile:         person.Mobile,
			Street1:        person.Street1,
			Street2:        person.Street2,
			Postcode:       person.Postcode,
			City:           person.City,
			Organisation15: p.organisationID,
			Institution15:  p.institutionID,
		})
		mems := itsl.ReadMembershipsForPerson(person.SyncPersonKey)
		for _, mem := range mems {
			p.memberships = append(p.memberships, itswizard_m_basic.Membership{
				PersonSyncKey:  mem.PersonID,
				GroupSyncKey:   mem.GroupID,
				Organisation15: p.organisationID,
				Institution15:  p.institutionID,
				Profile:        mem.Profile,
			})
			alleGruppenIds[mem.GroupID] = true
		}
	}
	for k, _ := range alleGruppenIds {
		outp := itsl.ReadGroup(k)
		p.groups = append(p.groups, itswizard_m_basic.Group{
			GroupSyncKey:   outp.Group.GroupSyncKey,
			Name:           outp.Group.Name,
			ParentGroupID:  outp.Group.ParentGroupID,
			IsCourse:       outp.Group.IsCourse,
			Organisation15: p.organisationID,
			Institution15:  p.institutionID,
		})
		if outp.Err != nil {
			log.Println(outp.Err)
			os.Exit(123)
		}
	}
	return p
}

func (p *Resync) SetDatafromOrganisation(rootGroup string) *Resync {
	itsl := itswizard_m_imses.NewImsesService(itswizard_m_imses.NewImsesServiceInput{
		Username: p.username,
		Password: p.password,
		Url:      p.endpoint,
	})

	for _, gr := range itsl.ReadAllGroup("0", rootGroup, p.organisationID, p.institutionID) {
		if gr.GroupSyncKey != rootGroup {
			p.groups = append(p.groups, gr)
		}
	}

	ptemp := itsl.ReadPersonsForGroup(rootGroup, p.organisationID, p.institutionID).Persons
	for i := 0; i < len(ptemp); i++ {
		if ptemp[i].PersonSyncKey != "" {
			if ptemp[i].PersonSyncKey != "itslearning_services" {
				if ptemp[i].PersonSyncKey != "itslearning_support" {
					p.persons = append(p.persons, ptemp[i])
				}
			}
		}
	}

	for _, k := range p.groups {
		mem, err, resp := itsl.ReadMembershipsForGroup(k.GroupSyncKey)
		if err != nil {
			log.Println(resp)
		}
		for i := 0; i < len(mem); i++ {
			if mem[i].PersonID != "" {
				if mem[i].PersonID != "itslearning_services" {
					if mem[i].PersonID != "itslearning_support" {
						p.memberships = append(p.memberships, itswizard_m_basic.Membership{
							PersonSyncKey:  mem[i].PersonID,
							GroupSyncKey:   mem[i].GroupID,
							Organisation15: p.organisationID,
							Institution15:  p.institutionID,
							Profile:        mem[i].Profile,
						})
					}
				}
			}
		}
	}

	mem, err, resp := itsl.ReadMembershipsForGroup(rootGroup)
	if err != nil {
		log.Println(resp)
	}
	for i := 0; i < len(mem); i++ {
		if mem[i].PersonID != "" {
			if mem[i].PersonID != "itslearning_services" {
				if mem[i].PersonID != "itslearning_support" {
					p.memberships = append(p.memberships, itswizard_m_basic.Membership{
						PersonSyncKey:  mem[i].PersonID,
						GroupSyncKey:   mem[i].GroupID,
						Organisation15: p.organisationID,
						Institution15:  p.institutionID,
						Profile:        mem[i].Profile,
					})
				}
			}
		}
	}
	return p
	/*
		alleGruppenIds := make(map[string]bool)
		out := itsl.ReadPersonsForGroup(rootGroup, p.organisationID, p.institutionID)
		for _, person := range out.Persons {
			if person.PersonSyncKey == "" {
				continue
			}
			p.persons = append(p.persons, itswizard_basic.Person{
				PersonSyncKey:  person.PersonSyncKey,
				FirstName:      person.FirstName,
				LastName:       person.LastName,
				Username:       person.Username,
				Profile:        person.Profile,
				Email:          person.Email,
				Phone:          person.Phone,
				Mobile:         person.Mobile,
				Street1:        person.Street1,
				Street2:        person.Street2,
				Postcode:       person.Postcode,
				City:           person.City,
				Organisation15: p.organisationID,
				Institution15:  p.institutionID,
			})
			mems := itsl.ReadMembershipsForPerson(person.PersonSyncKey)
			for _, mem := range mems {
				if mem.GroupID == "" {
					/*
					Schaue ob die GroupID teil der Rootgruppe ist!!
	*/
	/*
					continue
				}
				p.memberships = append(p.memberships, itswizard_basic.Membership{
					PersonSyncKey:  mem.PersonID,
					GroupSyncKey:   mem.GroupID,
					Organisation15: p.organisationID,
					Institution15:  p.institutionID,
					Profile:        mem.Profile,
				})
				alleGruppenIds[mem.GroupID] = true
			}
		}
		for k, _ := range alleGruppenIds {
			outp := itsl.ReadGroup(k)
			if outp.Group.GroupSyncKey == "" {
				continue
			}
			p.groups = append(p.groups, itswizard_basic.Group{
				GroupSyncKey:   outp.Group.GroupSyncKey,
				Name:           outp.Group.Name,
				ParentGroupID:  outp.Group.ParentGroupID,
				IsCourse:       outp.Group.IsCourse,
				Organisation15: p.organisationID,
				Institution15:  p.institutionID,
			})
			if outp.Err != nil {
				log.Println(outp)
				os.Exit(123)
			}
		}
		return p
	*/
}

func (p *Resync) GetSyncCache() (syncCache itswizard_m_syncCache.SyncCache) {
	syncCache.GroupsToImport = p.groups
	syncCache.PersonToImport = p.persons
	syncCache.MembershipToImport = p.memberships
	return
}

func (p *Resync) GetSyncCacheWithDatabaseDataAndSave(rootGroup string) {
	p.SetDatafromOrganisation(rootGroup)
	var sCache itswizard_m_syncCache.SyncCache
	p.DB.Find(&sCache.PersonToDelete)
	p.DB.Find(&sCache.MembershipToDelete)
	p.DB.Find(&sCache.GroupsToDelete)
	sCache.GroupsToImport = p.groups
	sCache.PersonToImport = p.persons
	sCache.MembershipToImport = p.memberships
	out, err := sCache.SaveCacheInJson(false)
	if err != nil {
		log.Println(err)
		return
	}
	err = p.DB.Save(&itswizard_m_syncCache.DbSyncCache{
		UserId:   0,
		Content:  out,
		Imported: false,
	}).Error
	if err != nil {
		log.Println(err)
		return
	}
}

func (p *Resync) GetSyncCacheWithDatabaseDataAndExcecute(rootGroup string) {
	p.SetDatafromOrganisation(rootGroup)
	var sCache itswizard_m_syncCache.SyncCache
	p.DB.Where("organisation15 = ?", p.organisationID).Find(&sCache.PersonToDelete)
	p.DB.Where("organisation15 = ?", p.organisationID).Find(&sCache.MembershipToDelete)
	p.DB.Where("organisation15 = ?", p.organisationID).Find(&sCache.GroupsToDelete)
	sCache.GroupsToImport = p.groups
	sCache.PersonToImport = p.persons
	sCache.MembershipToImport = p.memberships
	out, err := sCache.SaveCacheInJson(false)
	if err != nil {
		log.Println(err)
		return
	}
	_, logfile, err := sCache.Cache2Database(p.DB)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(logfile)
	err = p.DB.Save(&itswizard_m_syncCache.DbSyncCache{
		UserId:   0,
		Content:  out,
		Imported: true,
	}).Error
	if err != nil {
		log.Println(err)
		return
	}
}
